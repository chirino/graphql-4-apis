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

func TestNoContent(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
	}))
	defer func() { testServer.Close() }()

	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "no_content_test.json",
		},
		APIBase: apis.EndpointOptions{
			URL: testServer.URL,
		},
	})
	require.NoError(t, err)

	cxt := context.Background()
	result := ""
	err = engine.Exec(cxt, &result, `mutation{ noresult }`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"noresult":""}`, result)
}
