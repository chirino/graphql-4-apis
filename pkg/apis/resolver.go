package apis

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/chirino/graphql/inputconv"
	"github.com/chirino/graphql/qerrors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/chirino/graphql/resolvers"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

type Converter func(value reflect.Value, err error) (reflect.Value, error)

type resolver struct {
	resolvers         map[string]resolvers.Resolver
	resultConverters  map[string]Converter
	options           Config
	securityFunctions []func(query url.Values, headers http.Header, cookies []*http.Cookie) (url.Values, http.Header, []*http.Cookie)
	inputConverters   inputconv.TypeConverters
	next              resolvers.Resolver
}

func (resolver *resolver) convert(request *resolvers.ResolveRequest, next resolvers.Resolution) resolvers.Resolution {
	fieldType := request.Field.Type.String()
	if converter, ok := resolver.resultConverters[fieldType]; ok {
		return func() (value reflect.Value, err error) {
			return converter(next())
		}
	}
	return next
}

func (resolver *resolver) Resolve(request *resolvers.ResolveRequest, next resolvers.Resolution) resolvers.Resolution {
	key := request.ParentType.String() + ":" + request.Field.Name
	if r, ok := resolver.resolvers[key]; ok {
		resolution := r.Resolve(request, next)
		if resolution != nil {
			return resolver.convert(request, resolution)
		}
	}

	// We need these one to traverse the json results that are held as maps...
	resolution := resolvers.MapResolver.Resolve(request, next)
	if resolution != nil {
		return resolver.convert(request, resolution)
	}

	// And this one to handle Additional properties conversions.
	resolution = resolvers.FieldResolver.Resolve(request, next)
	if resolution != nil {
		return resolver.convert(request, resolution)
	}

	return next
}

func proxyHeaders(to http.Header, from *http.Request) {
	fromHeaders := from.Header
	for k, h := range fromHeaders {
		switch k {

		// Hop-by-hop headers... Don't forward these.
		// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
		case "Connection":
		case "Keep-Alive":
		case "Proxy-Authenticate":
		case "Proxy-Authorization":
		case "Te":
		case "Trailers":
		case "Transfer-Encoding":
		case "Upgrade":

		// Skip these headers which could affect our connection
		// to the upstream:
		case "Accept-Encoding":
		case "Sec-Websocket-Version":
		case "Sec-Websocket-Protocol":
		case "Sec-Websocket-Extensions":
		case "Sec-Websocket-Key":
		default:
			// Copy over any other headers..
			for _, header := range h {
				to.Add(k, header)
			}
		}
	}

	if clientIP, _, err := net.SplitHostPort(from.RemoteAddr); err == nil {
		if prior, ok := from.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		to.Set("X-Forwarded-For", clientIP)
	}

	if _, ok := from.Header["X-Forwarded-Host"]; !ok {
		if host := from.Header.Get("Host"); host != "" {
			to.Set("X-Forwarded-Host", host)
		}
	}

	if _, ok := from.Header["X-Forwarded-Proto"]; !ok {
		if from.TLS != nil {
			to.Set("X-Forwarded-Proto", "https")
		} else {
			to.Set("X-Forwarded-Proto", "http")
		}
	}
}

type DataLoadersKeyType string

const DataLoadersKey = DataLoadersKeyType("DataLoadersKey")

func (resolver *resolver) createLinkResolver(operation *openapi3.Operation, status []int, params map[string]string) resolvers.Func {
	return func(request *resolvers.ResolveRequest, next resolvers.Resolution) resolvers.Resolution {

		request.Args = map[string]interface{}{}
		for k, path := range params {
			keys := strings.Split(path, ".")
			value, err := navigateMap(request.Parent.Interface(), keys)
			if err != nil {
				return func() (reflect.Value, error) {
					return reflect.Value{}, errors.Wrapf(err, "could not get link argument at path: %s", path)
				}
			}
			request.Args[k] = fmt.Sprint(value)
		}

		dataLoaders := request.Context.Value(DataLoadersKey).(dataLoaders)
		marshaledArgs, err := json.Marshal(request.Args)
		if err != nil {
			return func() (reflect.Value, error) {
				return reflect.Value{}, errors.Wrapf(err, "could not create dataloader cache key")
			}
		}
		loadKey := loadKey{
			path:   operation.Extensions["path"].(string),
			method: operation.Extensions["method"].(string),
			args:   string(marshaledArgs),
		}

		loader := dataLoaders[loadKey]
		if loader == nil {
			loader = &CachedResolution{
				apply: resolver.resolve(request, operation, status),
			}
			dataLoaders[loadKey] = loader
		}
		return loader.resolution
	}
}

func navigateMap(value interface{}, keys []string) (interface{}, error) {
	current := reflect.ValueOf(value)
	current = resolvers.Dereference(current)

	for _, key := range keys {
		if current.Kind() != reflect.Map || current.Type().Key().Kind() != reflect.String {
			return nil, errors.New("can only navigate string keyed maps")
		}
		current = current.MapIndex(reflect.ValueOf(key))
	}

	return current.Interface(), nil
}

func (resolver resolver) resolve(gqlRequest *resolvers.ResolveRequest, operation *openapi3.Operation, expectedStatus []int) resolvers.Resolution {
	return func() (reflect.Value, error) {

		operationPath := operation.Extensions["path"].(string)
		operationMethod := operation.Extensions["method"].(string)

		query := url.Values{}
		headers := http.Header{}
		cookies := []*http.Cookie{}

		ctx := gqlRequest.Context
		if severRequest := ctx.Value("*net/http.Request"); severRequest != nil {
			if serverRequest, ok := severRequest.(*http.Request); ok {
				proxyHeaders(headers, serverRequest)

				cookie, err := serverRequest.Cookie("Authorization")
				if err == nil && cookie.Value != "" {
					headers.Set("Authorization", cookie.Value)
				}
			}
		}

		if resolver.options.APIBase.BearerToken != "" && headers.Get("Authorization") == "" {
			headers.Set("Authorization", "Bearer "+resolver.options.APIBase.BearerToken)
		}

		for _, param := range operation.Parameters {
			param := param.Value
			qlid := sanitizeName(param.Name)
			value, found := gqlRequest.Args[qlid]
			switch param.In {
			case "path":
				if !found { // all path params are required.
					panic("required path parameter not set: " + qlid)
				}
				operationPath = strings.ReplaceAll(operationPath, fmt.Sprintf("{%s}", param.Name), fmt.Sprintf("%v", value))

			case "query":
				if param.Required && !found {
					panic("required query parameter not set: " + qlid)
				}
				if found {
					query.Set(param.Name, fmt.Sprintf("%v", value))
				}
			case "header":
				if param.Name == "Accept-Encoding" {
					// the go http client automatically handles gzip decoding... manually setting the
					// the header disables this feature... so don't set it.
					continue
				}
				if param.Required && !found {
					panic("required header parameter not set: " + qlid)
				}
				if found {
					headers.Set(param.Name, fmt.Sprintf("%v", value))
				}

			case "cookie":
				cookies = append(cookies, &http.Cookie{
					Name:  param.Name,
					Value: fmt.Sprintf("%v", value),
				})
				// TODO: consider how to best handle these...
			}
		}

		headers.Set("Content-Type", "application/json")
		headers.Set("Accept", "application/json")

		apiURL, err := url.Parse(resolver.options.APIBase.URL)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}

		apiURL.Path += operationPath

		for _, f := range resolver.securityFunctions {
			query, headers, cookies = f(query, headers, cookies)
		}

		for _, cookie := range cookies {
			if v := cookie.String(); v != "" {
				headers.Add("Cookie", v)
			}
		}
		apiURL.RawQuery = query.Encode()

		client := resolver.options.APIBase.Client
		if client == nil {
			client = &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: resolver.options.APIBase.InsecureClient},
			}}
		}

		var body io.Reader = nil
		if operation.RequestBody != nil {
			content := operation.RequestBody.Value.Content.Get("application/json")
			if content != nil {

				v, err := resolver.inputConverters.Convert(gqlRequest.Field.Args.Get("body").Type, gqlRequest.Args["body"], "body")
				if err != nil {
					return reflect.Value{}, errors.WithStack(err)
				}

				data, err := json.Marshal(v)
				if err != nil {
					return reflect.Value{}, errors.WithStack(err)
				}

				body = bytes.NewReader(data)
			}
		}

		request, err := http.NewRequestWithContext(ctx, operationMethod, apiURL.String(), body)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}

		if operation.RequestBody != nil {
			content := operation.RequestBody.Value.Content.Get("application/json")
			if content != nil {
			}
		}

		request.Header = headers
		resp, err := client.Do(request)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		defer resp.Body.Close()

		for _, expected := range expectedStatus {
			if expected == resp.StatusCode {

				opResponse := operation.Responses.Get(resp.StatusCode)

				// Handle to NO_CONTENT case...
				if opResponse.Value.Content == nil {
					return reflect.ValueOf(""), nil
				}

				var result interface{}
				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					return reflect.Value{}, errors.WithStack(err)
				}
				return reflect.ValueOf(result), nil
			}
		}

		// All other statuses are considered errors...
		extensions := map[string]interface{}{}
		extensions["status"] = resp.StatusCode
		if resp.Header.Get("Content-Type") == "application/json" {
			response := map[string]interface{}{}
			err := json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return reflect.Value{}, err
			}
			extensions["response"] = response
		} else {
			all, _ := ioutil.ReadAll(resp.Body)
			extensions["response"] = string(all)
		}

		return reflect.Value{}, qerrors.Errorf("http response status code: %d", resp.StatusCode).WithExtensions(extensions).WithStack()
	}
}
