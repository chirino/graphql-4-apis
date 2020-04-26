package apis_test

import (
	"io/ioutil"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/require"
)

func TestLoadKubernetesAPI(t *testing.T) {
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "testdata/k8s.json",
		},
		APIBase: apis.EndpointOptions{
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
