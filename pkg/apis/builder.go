package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/chirino/graphql/inputconv"
	"github.com/chirino/graphql/resolvers"
	"github.com/chirino/graphql/schema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

func NewResolverFactory(doc *openapi3.T, options Config) (resolvers.Resolver, string, error) {
	builder := &builder{
		draft:          schema.New(),
		operationsById: map[string]*openapi3.Operation{},
		refCache:       map[string]interface{}{},
		resolver: &resolver{
			options: options,
			//next:             resolvers.DynamicResolverFactory(),
			resolvers:        make(map[string]resolvers.Resolver),
			resultConverters: make(map[string]Converter),
			inputConverters:  inputconv.TypeConverters{},
		},
	}

	for _, s := range doc.Security {
		for ssName := range s {
			ss := doc.Components.SecuritySchemes[ssName]
			if ss != nil && ss.Value != nil {
				switch ss.Value.Type {
				case "apiKey":
					if options.APIBase.ApiKey == "" {
						fmt.Println("API requires an api key, but it was not configured.")
						continue
					}
					switch ss.Value.In {
					case "header":
						builder.securityFunctions = append(builder.securityFunctions, func(query url.Values, headers http.Header, cookies []*http.Cookie) (url.Values, http.Header, []*http.Cookie) {
							headers.Set(ss.Value.Name, options.APIBase.ApiKey)
							return query, headers, cookies
						})
					case "query":
						builder.securityFunctions = append(builder.securityFunctions, func(query url.Values, headers http.Header, cookies []*http.Cookie) (url.Values, http.Header, []*http.Cookie) {
							query.Set(ss.Value.Name, options.APIBase.ApiKey)
							return query, headers, cookies
						})

					case "cookie":
						builder.securityFunctions = append(builder.securityFunctions, func(query url.Values, headers http.Header, cookies []*http.Cookie) (url.Values, http.Header, []*http.Cookie) {
							cookies = append(cookies, &http.Cookie{
								Name:  ss.Value.Name,
								Value: options.APIBase.ApiKey,
							})
							return query, headers, cookies
						})
					}
				}
			}
		}
	}

	// Lets index all the operations.. needed later when looking up operation due to links.
	for path, v := range doc.Paths {
		for method, operation := range v.Operations() {
			if operation.OperationID != "" {
				if builder.operationsById[operation.OperationID] != nil {
					// error?
					fmt.Println("Duplicate operation id found:", operation.OperationID)
				}
				if operation.Extensions == nil {
					operation.Extensions = map[string]interface{}{}
				}
				operation.Extensions["path"] = path
				operation.Extensions["method"] = method
				builder.operationsById[operation.OperationID] = operation
			}
		}
	}

	queryMethods := map[string]bool{"GET": true, "HEAD": true}
	for path, v := range doc.Paths {
		for method, operation := range v.Operations() {
			if queryMethods[method] {
				err := builder.addRootField(options.QueryType, operation)
				if err != nil {
					builder.options.Log.Printf("could not map api endpoint '%s %s': %s", method, path, err)
				}
			} else {
				err := builder.addRootField(options.MutationType, operation)
				if err != nil {
					builder.options.Log.Printf("could not map api endpoint '%s %s': %s", method, path, err)
				}
			}
		}
	}

	// Sort the type fields since we generated them by mutating..
	// which leads to then being in a random order based on the random order
	// they are received from the openapi doc.
	draft := builder.draft
	for _, t := range draft.Types {
		if t, ok := t.(*schema.Object); ok {
			sort.Slice(t.Fields, func(i, j int) bool {
				return t.Fields[i].Name < t.Fields[j].Name
			})
		}
		if t, ok := t.(*schema.InputObject); ok {
			sort.Slice(t.Fields, func(i, j int) bool {
				return t.Fields[i].Name < t.Fields[j].Name
			})
		}
	}

	if draft.Types[options.MutationType] != nil {
		draft.EntryPoints[schema.Mutation] = draft.Types[options.MutationType]
	}
	if draft.Types[options.QueryType] != nil {
		draft.EntryPoints[schema.Query] = draft.Types[options.QueryType]
	}

	if draft.Types["JSON"] != nil {
		builder.inputConverters["JSON"] = func(t schema.Type, value interface{}) (interface{}, error) {
			switch value := value.(type) {
			case string:
				return json.RawMessage(value), nil
			default:
				return nil, errors.New("unexpected type found for JSON scalar")
			}
		}
		builder.resultConverters["JSON"] = func(value reflect.Value, err error) (reflect.Value, error) {
			// input is an object, convert to a string
			if err != nil {
				return value, err
			}
			if value.IsNil() {
				return value, err
			}
			m := value.Interface().(map[string]interface{})
			if m == nil {
				return value, err
			}
			d, err := json.Marshal(m)
			if err != nil {
				return value, err
			}
			return reflect.ValueOf(string(d)), nil
		}
	}
	err := draft.ResolveTypes()
	if err != nil {
		return nil, "", err
	}
	return builder, draft.String(), nil
}

type builder struct {
	*resolver
	draft          *schema.Schema
	operationsById map[string]*openapi3.Operation
	refCache       map[string]interface{}
}

var _ resolvers.Resolver = &resolver{}

func (builder builder) addRootField(rootType string, operation *openapi3.Operation) error {

	draft := builder.draft

	var rootObject *schema.Object
	if t, ok := draft.Types[rootType]; ok {
		rootObject = t.(*schema.Object)
	} else {
		rootObject = &schema.Object{
			Name: rootType,
		}
		draft.Types[rootType] = rootObject
	}

	path := operation.Extensions["path"].(string)
	method := operation.Extensions["method"].(string)

	fieldName := sanitizeName(path)
	if operation.OperationID != "" {
		fieldName = sanitizeName(operation.OperationID)
	}

	if rootObject.Fields.Get(fieldName) != nil {
		builder.options.Log.Printf("field already exists: %s", fieldName)
		return nil
	}

	typePath := rootType + capitalizeFirstLetter(fieldName)

	qlType, status, err := builder.getOperationResponseType(operation, rootType, fieldName, typePath)
	if err != nil {
		builder.options.Log.Println(err.Error())
		return nil
	}

	text := ""
	if operation.Summary != "" {
		text = operation.Summary + "\n"
	}
	if operation.Description != "" {
		text = text + operation.Description + "\n"
	}
	text = text + "\n**endpoint:** `" + method + " " + path + "`"
	field := &schema.Field{
		Name: fieldName,
		Desc: desc(text),
		Type: qlType,
	}

	argNames := map[string]bool{}
	if operation.RequestBody != nil {
		content := operation.RequestBody.Value.Content.Get("application/json")
		if content != nil {
			fieldType, err := builder.addGraphQLType(content.Schema, typePath+"/body", true)
			if err != nil {
				builder.options.Log.Printf("dropping %s.%s field: required parameter '%s' type cannot be converted: %s", rootType, fieldName, "body", err)
				return nil
			}

			argName := makeUnique(argNames, "body")
			field.Args = append(field.Args, &schema.InputValue{
				Name: argName,
				Type: requiredType(fieldType, true),
			})
		}
	}

	if len(operation.Parameters) > 0 {
		for i, param := range operation.Parameters {

			if param.Value.In == "header" && param.Value.Name == "Accept-Encoding" {
				// the go http client automatically handles gzip decoding...
				// don't allow setting the Accept-Encoding header via a parameter.
				continue
			}

			argName := makeUnique(argNames, sanitizeName(param.Value.Name))
			fieldType, err := builder.addGraphQLType(getSchema(param.Value), fmt.Sprintf("%sArg%d", typePath, i), true)
			if err != nil {
				if param.Value.Required {
					builder.options.Log.Printf("dropping %s.%s field: required parameter '%s' type cannot be converted: %s", rootType, fieldName, param.Value.Name, err)
					return nil
				} else {
					builder.options.Log.Printf("dropping optional %s.%s field parameter: parameter '%s' type cannot be converted: %s", rootType, fieldName, param.Value.Name, err)
					continue
				}
			}

			field.Args = append(field.Args, &schema.InputValue{
				Desc: desc(param.Value.Description),
				Name: argName,
				Type: requiredType(fieldType, param.Value.Required),
			})
		}
	}

	rootObject.Fields = append(rootObject.Fields, field)
	builder.resolvers[rootType+":"+fieldName] = resolvers.Func(func(request *resolvers.ResolveRequest, next resolvers.Resolution) resolvers.Resolution {
		return builder.resolve(request, operation, status)
	})

	return nil
}

func (builder *builder) getOperationResponseType(operation *openapi3.Operation, rootType string, fieldName string, typePath string) (schema.Type, []int, error) {

	responseTypesToStatus := map[schema.Type][]int{}
	for statusText, response := range operation.Responses {
		status, err := strconv.Atoi(statusText)
		if err != nil {
			builder.options.Log.Printf("skipping %s.%s field response, not an integer: %s", rootType, fieldName, statusText)
		}
		if strings.HasPrefix(statusText, "2") {
			var qlType schema.Type = nil
			if response.Value.Content == nil {
				qlType = builder.NoContentType()
			} else {

				content := response.Value.Content.Get("application/json")
				if content != nil {

					qlType, err = builder.addGraphQLType(content.Schema, typePath, false)
					if err != nil {
						return nil, nil, errors.Errorf("dropping %s.%s field: result type cannot be converted: %s", rootType, fieldName, err)
					}
				}
			}

			if qlType != nil {
				statuses := responseTypesToStatus[qlType]
				if statuses == nil {
					responseTypesToStatus[qlType] = []int{status}
				} else {
					responseTypesToStatus[qlType] = append(statuses, status)
				}
			}
		}
	}

	switch len(responseTypesToStatus) {
	case 0:
		return nil, nil, errors.Errorf("dropping %s.%s field: graphql result type could not be determined", rootType, fieldName)
	case 1:
		for qlType, status := range responseTypesToStatus {
			return qlType, status, nil
		}
	}
	return nil, nil, errors.Errorf("dropping %s.%s field: graphql multiple result types not yet supported", rootType, fieldName)
}

func getSchema(value *openapi3.Parameter) *openapi3.SchemaRef {
	if value.Schema != nil {
		return value.Schema
	}
	if mediaType, ok := value.Content["application/json"]; ok {
		if mediaType.Schema != nil {
			return mediaType.Schema
		}
	}
	return nil
}

func (builder *builder) addGraphQLType(sf *openapi3.SchemaRef, path string, inputType bool) (schema.Type, error) {
	refCache := builder.refCache

	if sf == nil || sf.Value == nil {
		panic("a schema reference was not resolved.")
	}
	cacheKey := "o:" + sf.Ref
	if inputType {
		cacheKey = "i:" + sf.Ref
	}
	if sf.Ref != "" {
		if v, ok := refCache[cacheKey]; ok {
			if v, ok := v.(schema.Type); ok {
				return v, nil
			}
			return nil, v.(error)
		}
	}

	r, err := builder._addGraphQLType(sf, path, inputType)
	if err != nil {
		refCache[cacheKey] = err
		return nil, err
	}
	refCache[cacheKey] = r
	return r, nil
}

func (builder *builder) _addGraphQLType(sf *openapi3.SchemaRef, pathBasedTypeName string, inputType bool) (schema.Type, error) {
	draft := builder.draft

	typeName := pathBasedTypeName
	if sf.Ref != "" {
		typeName = strings.TrimPrefix(sf.Ref, "#/components/schemas/")
	}
	typeName = sanitizeName(typeName)
	pathBasedTypeName = typeName
	if inputType {
		typeName += "Input"
	} else {
		typeName += "Result"
	}

	switch sf.Value.Type {
	case "string":
		return draft.Types["String"], nil
	case "integer":
		return draft.Types["Int"], nil
	case "number":
		return draft.Types["Float"], nil
	case "boolean":
		return draft.Types["Boolean"], nil
	case "array":
		nestedType, err := builder.addGraphQLType(sf.Value.Items, pathBasedTypeName, inputType)
		if err != nil {
			return nil, err
		}
		return &schema.List{OfType: nestedType}, nil

	default: // Assume it's an object.

		// If object has no defined properties...
		if !hasProperties(sf.Value, map[*openapi3.Schema]bool{}) {

			// We can use a property wrapper if know the type of the values
			if sf.Value.AdditionalProperties != nil {
				nestedType, err := builder.addGraphQLType(sf.Value.AdditionalProperties, pathBasedTypeName, inputType)
				if err != nil {
					return nil, err
				}
				wrapper, err := builder.addPropWrapper(draft, nestedType, inputType)
				if err != nil {
					return nil, err
				}
				return &schema.List{OfType: &schema.NonNull{OfType: wrapper}}, nil
			}

			// I think it's safe to assume additional properties are allowed, if the object has no type
			t, found := draft.Types["JSON"]
			if !found {
				t = &schema.Scalar{
					Name: "JSON",
					Desc: desc("a JSON encoded object"),
				}
				draft.Types["JSON"] = t
			}
			return t, nil

		} else {
			if sf.Value.AdditionalProperties != nil {
				return nil, errors.New(fmt.Sprintf("cannot support additional prperties on graphql type '%s'", typeName))
			}
			if sf.Value.AdditionalPropertiesAllowed != nil && *sf.Value.AdditionalPropertiesAllowed {
				return nil, errors.New(fmt.Sprintf("cannot support additional prperties on graphql type '%s'", typeName))
			}
		}

		t := draft.Types[typeName]
		if t != nil {
			return t, nil
		}

		if inputType {
			t = &schema.InputObject{
				Desc: desc(sf.Value.Description),
				Name: typeName,
			}
		} else {
			t = &schema.Object{
				Desc: desc(sf.Value.Description),
				Name: typeName,
			}
		}
		// In case a type is recursive.. lets stick it in the cache now before we try to resolve it's fields..
		draft.Types[typeName] = t

		builder.addProperties(sf.Value, pathBasedTypeName, inputType, typeName, t)

		if inputType {
			object := t.(*schema.InputObject)
			if len(object.Fields) == 0 {
				delete(draft.Types, typeName)
				err := errors.New(fmt.Sprintf("graphql type '%s' would have no fields", typeName))
				return nil, err
			}
		} else {
			object := t.(*schema.Object)
			if len(object.Fields) == 0 {
				delete(draft.Types, typeName)
				err := errors.New(fmt.Sprintf("graphql type '%s' would have no fields", typeName))
				return nil, err
			}
		}

		return t, nil
	}

}

func hasProperties(value *openapi3.Schema, visited map[*openapi3.Schema]bool) bool {
	if visited[value] {
		return false
	}
	visited[value] = true
	if len(value.Properties) > 0 {
		return true
	}
	for _, ref := range value.OneOf {
		if hasProperties(ref.Value, visited) {
			return true
		}
	}
	for _, ref := range value.AnyOf {
		if hasProperties(ref.Value, visited) {
			return true
		}
	}
	for _, ref := range value.AllOf {
		if hasProperties(ref.Value, visited) {
			return true
		}
	}
	return false
}

func (builder *builder) addProperties(s *openapi3.Schema, pathBasedTypeName string, inputType bool, typeName string, graphqlType interface{}) {
	for _, sf := range s.AllOf {
		builder.addProperties(sf.Value, pathBasedTypeName, inputType, typeName, graphqlType)
	}
	for _, sf := range s.AnyOf {
		builder.addProperties(sf.Value, pathBasedTypeName, inputType, typeName, graphqlType)
	}
	for _, sf := range s.OneOf {
		builder.addProperties(sf.Value, pathBasedTypeName, inputType, typeName, graphqlType)
	}
	for name, ref := range s.Properties {
		fieldType, err := builder.addGraphQLType(ref, pathBasedTypeName+capitalizeFirstLetter(name), inputType)
		if err != nil {
			builder.options.Log.Printf("dropping openapi field '%s' from graphql type '%s': %s", name, typeName, err)
			continue
		}
		fieldName := sanitizeName(name)
		if inputType {
			object := graphqlType.(*schema.InputObject)
			newField := &schema.InputValue{
				Desc: desc(ref.Value.Description),
				Name: fieldName,
				Type: fieldType,
			}
			existingField := object.Fields.Get(fieldName)
			if existingField != nil {
				if !reflect.DeepEqual(newField.Type, existingField.Type) {
					builder.options.Log.Printf("field type conflict '%s.%s'", object.Name, fieldName)
				}
			} else {
				object.Fields = append(object.Fields, newField)
			}
		} else {
			object := graphqlType.(*schema.Object)
			newField := &schema.Field{
				Desc: desc(ref.Value.Description),
				Name: fieldName,
				Type: fieldType,
			}
			existingField := object.Fields.Get(fieldName)
			if existingField != nil {
				if !reflect.DeepEqual(newField.Type, existingField.Type) {
					builder.options.Log.Printf("field type conflict '%s.%s'", object.Name, fieldName)
				}
			} else {
				object.Fields = append(object.Fields, newField)
			}
		}
	}
	if !inputType {
		object := graphqlType.(*schema.Object)
		links := s.Extensions["x-links"]
		if linksRaw, ok := links.(json.RawMessage); ok {
			links := openapi3.Links{}
			err := json.Unmarshal(linksRaw, &links)
			if err != nil {
				builder.options.Log.Printf("invalid x-links extension: %s", string(linksRaw))
			} else {
				for field, link := range links {
					err := builder.addLink(object, field, pathBasedTypeName+capitalizeFirstLetter(field), link.Value)
					if err != nil {
						builder.options.Log.Printf("dropping x-links field '%s.%s': %v", typeName, field, err)
					}
				}
			}
		}
	}
}

func (builder *builder) addLink(qlType *schema.Object, fieldName string, typePath string, link *openapi3.Link) error {
	if qlType.Fields.Get(fieldName) != nil {
		return nil
	}

	// to avoid recursion...
	key := qlType.Name + "/" + fieldName
	if link.Extensions[key] != nil {
		return nil
	}
	link.Extensions[key] = true

	// link.Value.OperationID
	operation := builder.operationsById[link.OperationID]
	if operation == nil {
		return errors.New("could not find operation with id: " + link.OperationID)
	}

	responseType, status, err := builder.getOperationResponseType(operation, builder.options.QueryType, fieldName, typePath)
	if err != nil {
		return err
	}

	if qlType.Fields.Get(fieldName) != nil {
		return fmt.Errorf("could not add link named '%s': field already exists", fieldName)
	}
	qlType.Fields = append(qlType.Fields, &schema.Field{
		Desc: desc(link.Description),
		Name: fieldName,
		Type: responseType,
	})

	params := map[string]string{}
	for k, v := range link.Parameters {
		params[k] = fmt.Sprint(v)
	}

	builder.resolvers[qlType.Name+":"+fieldName] = builder.createLinkResolver(operation, status, params)

	return nil
}

func (builder *builder) addPropWrapper(draft *schema.Schema, nestedType schema.Type, inputType bool) (schema.NamedType, error) {

	nestedTypeLong := schema.DeepestType(nestedType).String()
	graphTypeName := nestedTypeLong + "ResultProp"
	if inputType {
		graphTypeName = nestedTypeLong + "InputProp"
	}

	// Wrapper type might already exist...
	graphType := draft.Types[graphTypeName]
	if graphType != nil {
		return graphType, nil
	}

	if inputType {

		graphType = &schema.InputObject{
			Name: graphTypeName,
			Fields: []*schema.InputValue{
				{
					Name: "key",
					Type: requiredType(draft.Types["String"], true),
				},
				{
					Name: "value",
					Type: nestedType,
				},
			},
		}

		builder.inputConverters["["+graphType.TypeName()+"!]"] = func(t schema.Type, value interface{}) (interface{}, error) {
			if value == nil {
				return nil, nil
			}
			// input is an array.. convert to a map...
			if value, ok := value.([]interface{}); ok {
				result := make(map[string]interface{}, len(value))
				for _, item := range value {
					if item, ok := item.(map[string]interface{}); ok {
						if key, ok := item["key"].(string); ok {
							value := item["value"]
							result[key] = value
						} else {
							return nil, errors.Errorf("input conversion of "+t.String()+" type failed: expected array item key to be a string, got: %T", key)
						}
					} else {
						return nil, errors.Errorf("input conversion of "+t.String()+" type failed: expected array item to be a map, got: %T", item)
					}
				}
				return result, nil
			}
			return nil, errors.Errorf("input conversion of "+t.String()+" type failed: expected array, got: %T", value)
		}

	} else {
		graphType = &schema.Object{
			Desc: desc("A property entry"),
			Name: graphTypeName,
			Fields: []*schema.Field{
				{
					Name: "key",
					Type: requiredType(draft.Types["String"], true),
				},
				{
					Name: "value",
					Type: nestedType,
				},
			},
		}

		builder.resultConverters["["+graphType.TypeName()+"!]"] = func(value reflect.Value, err error) (reflect.Value, error) {
			// input is a map... convert to an array
			if err != nil {
				return value, err
			}
			if value.IsNil() {
				return value, err
			}
			m := value.Interface().(map[string]interface{})
			if m == nil {
				return value, err
			}

			type Prop struct {
				Key   string      `json:"key"`
				Value interface{} `json:"value"`
			}

			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var props []Prop
			for _, k := range keys {
				props = append(props, Prop{Key: k, Value: m[k]})
			}
			return reflect.ValueOf(props), nil
		}
	}

	draft.Types[graphType.TypeName()] = graphType
	return graphType, nil
}

func (builder *builder) NoContentType() schema.Type {
	draft := builder.draft
	t := draft.Types["NO_CONTENT"]
	if t == nil {
		t = &schema.Scalar{
			Name:       "NO_CONTENT",
			Desc:       desc("An empty result"),
			Directives: nil,
		}
		draft.Types["NO_CONTENT"] = t
	}
	return t
}
