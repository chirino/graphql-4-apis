package tests_test

import (
	"bytes"
	"encoding/json"
	"github.com/chirino/graphql"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	m := mux.NewRouter()
	m.HandleFunc("/animals", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(
			[]map[string]string{
				{"id": "mickey"},
				{"id": "joe"},
				{"id": "mighty"},
			})
	})
	m.HandleFunc("/animals/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch id {
		case "mickey":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":             id,
				"animal_type_id": "mouse",
				"name":           "Mickey Mouse",
			})
		case "joe":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":             id,
				"animal_type_id": "human",
				"name":           "Joe Johnson",
			})
		case "mighty":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":             id,
				"animal_type_id": "mouse",
				"name":           "Mighty Mouse",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	mouseHitCounter := 0
	humanHitCounter := 0
	m.HandleFunc("/animal_types/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch id {
		case "mouse":
			mouseHitCounter += 1
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":      id,
				"species": "Mouse",
			})
		case "human":
			humanHitCounter += 1
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":      id,
				"species": "Human",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	server := httptest.NewServer(m)
	defer server.Close()

	messages := bytes.NewBuffer(nil)
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "link_test.yaml",
		},
		APIBase: apis.EndpointOptions{
			URL: server.URL,
		},
		Log: log.New(messages, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()

	if os.ExpandEnv("${GENERATE_TEST_GRAPHQL_FILES}") == "true" {
		ioutil.WriteFile("link_test.graphql", []byte(actual), 0644)
	}
	AssertEquals(t, "", messages.String())

	file, err := ioutil.ReadFile("link_test.graphql")
	require.NoError(t, err)
	expected := string(file)
	AssertEquals(t, expected, actual)

	response := engine.ServeGraphQL(&graphql.Request{
		Query: `{
			getAnimalIds {
				id
				animal {
					name
					animal_type {
                  		species
					}
				}
			}
		}`,
	})
	require.Empty(t, response.Errors)
	actual = string(response.Data)
	AssertEquals(t, `{"getAnimalIds":[{"id":"mickey","animal":{"name":"Mickey Mouse","animal_type":{"species":"Mouse"}}},{"id":"joe","animal":{"name":"Joe Johnson","animal_type":{"species":"Human"}}},{"id":"mighty","animal":{"name":"Mighty Mouse","animal_type":{"species":"Mouse"}}}]}`, actual)

	// Verify that we don't repeat requests for the same resource multiple times.
	AssertEquals(t, mouseHitCounter, 1)
	AssertEquals(t, humanHitCounter, 1)

}
