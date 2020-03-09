package api_test

import (
	"context"
	"github.com/chirino/graphql"
	"github.com/chirino/graphql-4-apis/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdditionProperties(t *testing.T) {
	engine := graphql.New()

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"a":["1"],"b":["2"]}`, string(body))
		res.Write([]byte(`{"a":["2"],"b":["4"]}`))
	}))
	defer func() { testServer.Close() }()

	err := api.MountApi(engine, api.ApiResolverOptions{
		Openapi: api.EndpointOptions{
			URL: "testdata/additionalProperties.json",
		},
		APIBase: api.EndpointOptions{
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
