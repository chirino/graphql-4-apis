package apis_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdditionProperties(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"a":["1"],"b":["2"]}`, string(body))
		res.Write([]byte(`{"a":["2"],"b":["4"]}`))
	}))
	defer func() { testServer.Close() }()

	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "testdata/additionalProperties.json",
		},
		APIBase: apis.EndpointOptions{
			URL: testServer.URL,
		},
	})
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
	assert.Equal(t, `{"action":[{"key":"a","value":["2"]},{"key":"b","value":["4"]}]}`, result)
}