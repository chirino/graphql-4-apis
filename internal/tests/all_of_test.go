package tests

import (
	"context"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllOf(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte(`{"age":21}`))
	}))
	defer func() { testServer.Close() }()

	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "all_of_test.json",
		},
		APIBase: apis.EndpointOptions{
			URL: testServer.URL,
		},
	})
	require.NoError(t, err)

	//fmt.Println(engine.Schema)

	cxt := context.Background()
	result := ""
	err = engine.Exec(cxt, &result, `mutation{ action { age } }`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"action":{"age":21}}`, result)
}
