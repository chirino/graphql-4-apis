package cmd

import (
	_ "github.com/chirino/graphql-4-apis/internal/cmd/new"
	graphql_4_apis "github.com/chirino/graphql-4-apis/internal/cmd/root"
	_ "github.com/chirino/graphql-4-apis/internal/cmd/serve"
)

func Main() {
	graphql_4_apis.Main()
}
