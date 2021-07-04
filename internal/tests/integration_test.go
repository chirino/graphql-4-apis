package tests_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeatherDotCom(t *testing.T) {

	// Get the API key from https://callforcode.weather.com/admin/my-api-key/
	apiKey, found := os.LookupEnv("WEATHER_DOT_COM_APIKEY")
	if !found {
		t.Skip("set the WEATHER_DOT_COM_APIKEY env variable if you want to run integration tests against a weather.com")
	}

	engine, err := apis.CreateGatewayEngine(apis.Config{
		Openapi: apis.EndpointOptions{
			URL: "https://weather.com/swagger-docs/sun/v1/sunV1DailyForecast.json",
		},
		APIBase: apis.EndpointOptions{
			ApiKey: apiKey,
		},
		Log: log.New(ioutil.Discard, "", 0),
	})
	require.NoError(t, err)
	err = engine.Schema.Parse(`
        schema {
            query: Query
            mutation: Mutation
        }
    `)

	result := json.RawMessage{}
	err = engine.Exec(nil, &result, `
query {
  getSunDailyForecastByLocation(postalCode:"33548:4:US", language:"en-US",days:3) {
    forecasts {
      dow
      lunar_phase
    }
  }
}`)
	require.NoError(t, err)
	assert.Contains(t, string(result), `{"getSunDailyForecastByLocation":{"forecasts":[{`)
}
