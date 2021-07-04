package tests_test

import (
	"context"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdditionProperties(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"a":["1"],"b":["2"]}`, string(body))
		_, err = res.Write([]byte(`{"a":["2"],"b":["4"]}`))
		require.NoError(t, err)
	}))
	defer func() { testServer.Close() }()

	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "additional_properties_with_type_test.json",
		},
		APIBase: apis.EndpointOptions{
			URL: testServer.URL,
		},
	})
	require.NoError(t, err)
	err = engine.Schema.Parse(`
        schema {
            mutation: Mutation
        }
    `)
	require.NoError(t, err)

	cxt := context.Background()
	result := ""
	err = engine.Exec(cxt, &result, `mutation{ action(body:[{key:"a", value:["1"]}, {key:"b", value:["2"]}]) { key, value } }`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"action":[{"key":"a","value":["2"]},{"key":"b","value":["4"]}]}`, result)
}
