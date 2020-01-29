# GraphQL-4-APIs

Lets you create a [GraphQL](https://graphql.org/) gateway service for your
[OpenAPI Specifications (OAS)](https://github.com/OAI/OpenAPI-Specification) 
or [Swagger](https://swagger.io/specification/v2/) documented API services.  This lets you execute
GraphQL queries/mutations against your APIs.

## Getting started

### CLI

The Command Line Interface (CLI) provides a convenient way to start a GraphQL server wrapping an API for a given OpenAPI Specification:

1. Install the `graphql-4-apis` CLI using:
   ```bash
   go get -u github.com/chirino/graphql-4-apis
   ```
2. Then, run the OpenAPI-to-GraphQL command and point it to an OpenAPI Specification:
   ```bash
   $ graphql-4-apis api myopenapi.json
   GraphQL service running at http://localhost:8080/graphql
   GraphiQL UI running at http://localhost:8080/graphiql
   ```

## Characteristics

- **Queries**
  
  Defined headers, query, and path open api parameters become type checked parameters to the GraphQL queries.  Responses
  get automatically generated graphql types.

- **Mutations**

  Non-safe, non-idempotent API operations are translated to GraphQL [mutations](http://graphql.org/learn/queries/#mutations). 
  GraphQL Input Objects schemas are generated for the input body so that that input data is type checked.

- **Authentication**

  All HTTP headers on the GraphQL request are passed through to the upstream API.
  
  A login mutation is provided so that allows you to store a
  bearer token in an http cookie that is used in subsequent request to populate the Authorization header 
  on API  requests.  This comes in handy if your using the GraphiQL interface since you can't 
  easily set Authorization headers using that interface.  A logout mutation is also available to
  clear the cookie.
  
- **API Sanitation**

  API names not compatible with GraphQL are automatically sanitized. For example, API parameters and data definition names with unsupported 
  characters (e.g., `-`, `.`, `,`, `:`, `;`...) are replaced with `_`.

- **Swagger and OpenAPI 3 support** OpenAPI-to-GraphQL can handle both Swagger (OpenAPI specification 2.0) as well as OpenAPI specification 3.

- **Support for json objects with dynamic keys** GraphQL object types requires all fields of a type to be known, openapi
allows json types with dynamic object keys.  In these cases, we map the object type to an array of key value pairs `[<ValueType>ResultProp!]` 
that using this template:
```graphql
type <ValueType>ResultProp {
    key String!
    value <ValueType> 
}
```

## License

[BSD](./LICENSE)

## Development

* We love [pull requests](https://github.com/chirino/graphql-4-apis/pulls)
* Is one of your APIs not working well with this project?  [Lets us know](https://github.com/chirino/graphql-4-apis/issues)
* GraphQL-4-APIs is written in [Go](https://golang.org/).  It should work on any platform where go is supported.
* Built on this [GraphQL](https://github.com/chirino/graphql) framework
* Currently focused on being a CLI based server, but eventually would like to provided library to allow it to be used 
  in your custom GraphQL servers

## Future Work

* make operation id's optional
* use links in the openapi v3 spec to resolve relationships in the the graph structure
* support passing custom headers and query parameters to api calls
* support subscriptions
* allow customizing the generated GraphQL schema
* https secured graphql endpoints 
* sanitization on json schema field names.
* setup ci jobs and have some binary releases

## Similar projects

- [openapi-to-graphql](https://github.com/IBM/openapi-to-graphql) Very similar to this project, written in javascript

- [swagger-to-graphql](https://github.com/yarax/swagger-to-graphql) turns a given Swagger (OpenAPI Specification 2.0) into a GraphQL interface, which resolves against the original API. GraphQL schema is based on endpoints, not on data definitions. No links are considered.

- [json-to-graphql](https://github.com/aweary/json-to-graphql) turns given JSON objects / arrays into a GraphQL schema. `resolve` functions need to be provided by the user.

- [StackOverflow discussion](https://stackoverflow.com/questions/38339442/json-schema-to-graphql-schema-converters) points to the above projects.

