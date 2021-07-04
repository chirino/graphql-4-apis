package tests_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/require"
)

func TestConnectorMgmtAPI(t *testing.T) {
	messages := bytes.NewBuffer(nil)
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "connector_mgmt_test.yaml",
		},
		APIBase: apis.EndpointOptions{
			URL:    "http://fake:8080",
			ApiKey: "fake",
		},
		Log: log.New(messages, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	if os.ExpandEnv("${GENERATE_TEST_GRAPHQL_FILES}") == "true" {
		ioutil.WriteFile("connector_mgmt_test.graphql", []byte(actual), 0644)
	}
	AssertEquals(t, "", messages.String())

	file, err := ioutil.ReadFile("connector_mgmt_test.graphql")
	require.NoError(t, err)
	expected := string(file)

	AssertEquals(t, expected, actual)
}
