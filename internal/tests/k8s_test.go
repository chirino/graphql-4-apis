package tests_test

import (
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/require"
)

func TestLoadKubernetesAPI(t *testing.T) {
	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "k8s_test.json",
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
		ioutil.WriteFile("k8s_test.graphql", []byte(actual), 0644)
	}

	file, err := ioutil.ReadFile("k8s_test.graphql")
	require.NoError(t, err)
	expected := string(file)

	AssertEquals(t, expected, actual)
}

func AssertEquals(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(fmt.Sprint(expected)),
			B:        difflib.SplitLines(fmt.Sprint(actual)),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})
		t.Errorf("actual does not match expected, diff:\n%s\n", diff)
	}
}
