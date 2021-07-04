package tests_test

import (
	"bytes"
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

func TestOneOfWithDiscriminator(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		AssertEquals(t, `{}`, string(data))
		w.Write([]byte(`{
			"kind": "human",
			"address": "Florida"
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	messages := bytes.NewBuffer(nil)
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "testdata/one_of_with_discriminator.yaml",
		},
		APIBase: apis.EndpointOptions{
			URL: server.URL,
		},
		Log: log.New(messages, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	ioutil.WriteFile("testdata/one_of_with_discriminator.graphql", []byte(actual), 0644)

	AssertEquals(t, "", messages.String())

	file, err := ioutil.ReadFile("testdata/one_of_with_discriminator.graphql")
	require.NoError(t, err)
	expected := string(file)
	AssertEquals(t, expected, actual)

	response := engine.ServeGraphQL(&graphql.Request{
		Query: `mutation{
			example(body:{kind:"dog", owner:"nick"})
		}`,
	})
	require.Empty(t, response.Errors)
	actual = string(response.Data)
	AssertEquals(t, `{"example":{}}`, actual)

}
