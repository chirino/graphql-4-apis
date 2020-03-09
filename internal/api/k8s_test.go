package api_test

import (
	"github.com/chirino/graphql"
	"github.com/chirino/graphql-4-apis/internal/api"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestLoadKubernetesAPI(t *testing.T) {
	engine := graphql.New()

	err := api.MountApi(engine, api.ApiResolverOptions{
		Openapi: api.EndpointOptions{
			URL: "testdata/k8s.json",
		},
		APIBase: api.EndpointOptions{
			URL: "http://fake:8080",
		},
		Logs: ioutil.Discard,
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	//ioutil.WriteFile("testdata/k8s.graphql", []byte(actual), 0644)

	file, err := ioutil.ReadFile("testdata/k8s.graphql")
	require.NoError(t, err)
	expected := string(file)

	require.Equal(t, actual, expected)
}
