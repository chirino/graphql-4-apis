package tests_test

import (
	"bytes"
	"encoding/json"
	"github.com/chirino/graphql"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdditionalProperties(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m := map[string]string{}
		err = json.Unmarshal(data, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m["hello"] = "world"
		err = json.NewEncoder(w).Encode(m)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	messages := bytes.NewBuffer(nil)
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "testdata/additional_properties.yaml",
		},
		APIBase: apis.EndpointOptions{
			URL: server.URL,
		},
		Log: log.New(messages, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	// ioutil.WriteFile("testdata/additional_properties.graphql", []byte(actual), 0644)

	AssertEquals(t, "", messages.String())

	file, err := ioutil.ReadFile("testdata/additional_properties.graphql")
	require.NoError(t, err)
	expected := string(file)
	AssertEquals(t, expected, actual)

	response := engine.ServeGraphQL(&graphql.Request{
		Query: `mutation{
			example(body:"{\"test\":\"request\"}")
		}`,
	})
	require.Empty(t, response.Errors)
	actual = string(response.Data)
	AssertEquals(t, `{"example":"{\"hello\":\"world\",\"test\":\"request\"}"}`, actual)

}
