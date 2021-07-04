package tests_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/require"
)

func TestLoadOpenshiftAPI(t *testing.T) {
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "openshift_test.json",
		},
		APIBase: apis.EndpointOptions{
			URL:    "http://fake:8080",
			ApiKey: "fake",
		},
		Log: log.New(ioutil.Discard, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	if os.ExpandEnv("${GENERATE_TEST_GRAPHQL_FILES}") == "true" {
		ioutil.WriteFile("openshift_test.graphql", []byte(actual), 0644)
	}
	file, err := ioutil.ReadFile("openshift_test.graphql")
	require.NoError(t, err)
	expected := string(file)

	AssertEquals(t, expected, actual)
}
