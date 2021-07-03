package apis_test

import (
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"io/ioutil"
	"log"
	"reflect"
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
			URL:    "http://fake:8080",
			ApiKey: "fake",
		},
		Log: log.New(ioutil.Discard, "", 0),
	})
	require.NoError(t, err)

	actual := engine.Schema.String()
	//ioutil.WriteFile("testdata/k8s.graphql", []byte(actual), 0644)

	file, err := ioutil.ReadFile("testdata/k8s.graphql")
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
