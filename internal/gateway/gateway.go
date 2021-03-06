package gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/httpgql"
)

func ListenAndServe(config apis.Config) {
	engine, err := apis.CreateGatewayEngine(config)
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

	addr := ":8080"
	http.Handle("/graphql", &httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	http.Handle("/", graphiql.New("ws://localhost"+addr+"/graphql", true))
	fmt.Println("GraphQL service running at http://localhost" + addr + "/graphql")
	fmt.Println("GraphiQL UI running at http://localhost" + addr + "/")
	log.Fatal(http.ListenAndServe(addr, nil))
}
