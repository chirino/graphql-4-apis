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

- **Data-centric**
  The GraphQL interface is created around the data definitions in the given OAS, not around the endpoints, leading to a natural use of GraphQL.

  <img src="https://raw.githubusercontent.com/ibm/openapi-to-graphql/master/docs/data-centric.png" alt="Example of data-centric design" width="600">

- **Automatic query resolution**
  Automatically generated resolvers translate (nested) GraphQL queries to API requests. Request results are translated back to GraphQL responses.


- **Mutations**
  Non-safe, non-idempotent API operations (e.g., `POST`, `PUT`, `DELETE`) are translated to GraphQL [mutations](http://graphql.org/learn/queries/#mutations). Input payload is type-checked.

  <img src="https://raw.githubusercontent.com/ibm/openapi-to-graphql/master/docs/mutations.png" alt="Example of mutation" width="600">

- **Authentication**

  All HTTP headers on the GraphQL request are passed through to the upstream API.
  
  A login mutation is provided so that allows you to store a
  bearer token in an htpt cookie that is used in subsequent request to populate the Authorization header 
  on API  requests.  This comes in handy if your using the GraphiQL interface since you can't 
  easily set Authorization headers using that interface.  A logout mutation is also available to
  clear the cookie.
  
- **API Sanitation**
  Parts of an API that not compatible with GraphQL are automatically sanitized. For example, API parameters and data definition names with unsupported characters (e.g., `-`, `.`, `,`, `:`, `;`...) are removed. GraphQL queries are desanitized to correctly invoke the REST API and the responses are resanitized to create GraphQL-compliant results.

  <img src="https://raw.githubusercontent.com/ibm/openapi-to-graphql/master/docs/sanitization.png" alt="Example of sanitation" width="300">

- **Swagger and OpenAPI 3 support** OpenAPI-to-GraphQL can handle both Swagger (OpenAPI specification 2.0) as well as OpenAPI specification 3.

## Future Work

* make operation id's optional
* use links in the openapi v3 spec to resolve relationships in the the graph structure
* support passing custom headers and query parameters to api calls
* support subscriptions
* allow customizing the generated GraphQL schema
* https secured graphql endpoints 

## License

[BSD](./LICENSE)

## Development

GraphQL-4-APIs is written in [Go](https://golang.org/).  We love pull requests.

## Similar projects

- [openapi-to-graphql](https://github.com/IBM/openapi-to-graphql)

- [swagger-to-graphql](https://github.com/yarax/swagger-to-graphql) turns a given Swagger (OpenAPI Specification 2.0) into a GraphQL interface, which resolves against the original API. GraphQL schema is based on endpoints, not on data definitions. No links are considered.

- [json-to-graphql](https://github.com/aweary/json-to-graphql) turns given JSON objects / arrays into a GraphQL schema. `resolve` functions need to be provided by the user.

- [StackOverflow discussion](https://stackoverflow.com/questions/38339442/json-schema-to-graphql-schema-converters) points to the above projects.

