module github.com/chirino/graphql-4-apis

require (
	github.com/chirino/graphql v0.0.0-20200217144534-2b17ff897cbb
	github.com/getkin/kin-openapi v0.2.1-0.20200126120519-bd363bbcb48f
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.7.4
	github.com/kr/text v0.1.0
	github.com/pkg/errors v0.8.0
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.3.0
)

go 1.13

replace github.com/chirino/graphql => ../graphql

replace github.com/getkin/kin-openapi => ../kin-openapi
