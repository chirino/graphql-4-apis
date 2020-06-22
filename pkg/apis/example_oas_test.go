package apis_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/chirino/graphql-4-apis/internal/dom"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/httpgql"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestExampleOasAPI(t *testing.T) {

	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "testdata/example_oas.json",
		},
		APIBase: apis.EndpointOptions{
			URL: "http://localhost:8080/api",
		},
		Log: log.New(ioutil.Discard, "", 0),
	})
	require.NoError(t, err)
	err = engine.Schema.Parse(`
        schema {
            query: Query
            mutation: Mutation
        }
    `)

	actual := engine.Schema.String()
	ioutil.WriteFile("testdata/example_oas.graphql", []byte(actual), 0644)

	file, err := ioutil.ReadFile("testdata/example_oas.graphql")
	require.NoError(t, err)
	expected := string(file)

	require.Equal(t, actual, expected)

	f, err := os.Open("testdata/example_oas_data.json")
	require.NoError(t, err)
	data := dom.New()
	err = json.NewDecoder(f).Decode(&data)
	require.NoError(t, err)

	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/graphql", &httpgql.Handler{ServeGraphQLStream: engine.ServeGraphQLStream})
	router.Handle("/graphiql", graphiql.New("ws://localhost:8080/graphql", true))

	encode := func(w http.ResponseWriter, status int, r interface{}) {
		if r == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode("Not Found")
		} else {
			w.WriteHeader(status)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(r)
		}
	}

	router.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		encode(w, 202, data.GetDom("users").Values())
	})

	router.HandleFunc("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		encode(w, 200, data.GetDom("users", id))
	})
	server := &http.Server{Addr: ":8080", Handler: router}
	go func() {
		server.ListenAndServe()
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Press Enter to Exit: ")
	reader.ReadString('\n')

	server.Close()
}
