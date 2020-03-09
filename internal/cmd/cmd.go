package cmd

import (
	_ "github.com/chirino/graphql-4-apis/internal/cmd/api"
	_ "github.com/chirino/graphql-4-apis/internal/cmd/config"
	graphql_4_apis "github.com/chirino/graphql-4-apis/internal/cmd/graphql-4-apis"
)

func Main() {
	graphql_4_apis.Main()
}
