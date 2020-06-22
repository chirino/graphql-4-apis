package apis_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/require"
)

func TestLoadOpenshiftAPI(t *testing.T) {
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "testdata/openshift.json",
		},
		APIBase: apis.EndpointOptions{
			URL:    "http://fake:8080",
			ApiKey: "fake",
		},
		Log: log.New(ioutil.Discard, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	// ioutil.WriteFile("testdata/openshift.graphql", []byte(actual), 0644)

	file, err := ioutil.ReadFile("testdata/openshift.graphql")
	require.NoError(t, err)
	expected := string(file)

	require.Equal(t, actual, expected)
}
