package gateway

import (
    "fmt"
    "github.com/chirino/graphql"
    "github.com/chirino/graphql-4-apis/internal/api"
    "github.com/chirino/graphql/relay"
    "github.com/friendsofgo/graphiql"
    "log"
    "net/http"
)

func ListenAndServe(config api.ApiResolverOptions) {
    engine := graphql.New()
    err := api.MountApi(engine, config)
    if err != nil {
        log.Fatalf("%+v", err)
    }

    err = engine.Schema.Parse(`
        type Query {
            # Access to the API
            api: QueryApi,
        }
        type Mutation {
            # Saves a Authorization Bearer token in a browser cookie that 
            # is then subsequently used when issuing requests to the API.
            login(token:String!): String
            # Clears the Authorization Bearer token previously stored in a browser cookie.
            logout(): String
            # Access to the API
            api: MutationApi,
        }
        schema {
            query: Query
            mutation: Mutation
        }
    `)
    if err != nil {
        log.Fatalf("%+v", err)
    }
    engine.Root = root(0)

    http.Handle("/graphql", &relay.Handler{Engine: engine})
    graphiql, _ := graphiql.NewGraphiqlHandler("/graphql")
    http.Handle("/graphiql", graphiql)

    addr := ":8080"
    fmt.Println("GraphQL service running at http://localhost" + addr + "/graphql")
    fmt.Println("GraphiQL UI running at http://localhost" + addr + "/graphiql")
    log.Fatal(http.ListenAndServe(addr, nil))
}

