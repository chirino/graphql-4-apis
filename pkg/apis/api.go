package apis

import (
	"context"
	errors "errors"
	"github.com/chirino/graphql/schema"
	"net/http"
	"os"

	"log"

	"github.com/chirino/graphql"
	"github.com/chirino/graphql/resolvers"
)

// EndpointOptions defines how to access an endpoint URL
type EndpointOptions struct {
	// URL is the endpoint or endpoint base path that will be accessed.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// BearerToken is the Authentication Bearer token that will added to the request headers.
	BearerToken string `yaml:"bearer-token,omitempty" json:"bearer-token,omitempty"`
	// Api key for the endpoint.
	ApiKey string `yaml:"api-key,omitempty" json:"api-key,omitempty"`
	// InsecureClient allows the client request to connect to TLS servers that do not have a valid certificate.
	InsecureClient bool         `yaml:"insecure-client,omitempty" json:"insecure-client,omitempty"`
	Client         *http.Client `yaml:"-" json:"-"`
	// OpenapiDocument is the content of the openapi document, if empty it will be loaded from the URL
	OpenapiDocument []byte `yaml:"-" json:"-"`
}

type Config struct {
	Openapi      EndpointOptions `json:"spec,omitempty",yaml:"spec,omitempty"`
	APIBase      EndpointOptions `json:"api,omitempty",yaml:"api,omitempty"`
	QueryType    string
	MutationType string
	Log          *log.Logger
}

func CreateGatewayEngine(option Config) (*graphql.Engine, error) {
	engine := graphql.New()
	o := Config{
		QueryType:    "Query",
		MutationType: "Mutation",
	}
	if option.Log != nil {
		o.Log = option.Log
	} else {
		o.Log = log.New(os.Stderr, "graphql-4-apis: ", 0)
	}
	if option.QueryType != "" {
		o.QueryType = option.QueryType
	}
	if option.MutationType != "" {
		o.MutationType = option.MutationType
	}
	o.Openapi = option.Openapi
	o.APIBase = option.APIBase

	doc, err := LoadOpenApiV2orV3Doc(o.Openapi)
	if err != nil {
		return nil, err
	}

	// If the APIBase.URL is not configured.. try to figure it out from the openapi doc...
	if o.APIBase.URL == "" {
		for _, server := range doc.Servers {
			if server != nil && server.URL != "" {
				o.APIBase.URL = server.URL
				break
			}
		}
	}

	if o.APIBase.URL == "" {
		return nil, errors.New("api base URL is not configured")
	}

	resolver, schemaText, err := NewResolverFactory(doc, o)
	if err != nil {
		return nil, err
	}
	err = engine.Schema.Parse(schemaText)
	if err != nil {
		return nil, err
	}

	engine.OnRequestHook = func(r *graphql.Request, doc *schema.QueryDocument, op *schema.Operation) error {
		r.Context = context.WithValue(r.GetContext(), DataLoadersKey, dataLoaders{})
		return nil
	}

	engine.Resolver = resolvers.List(resolver, engine.Resolver)
	return engine, nil
}
