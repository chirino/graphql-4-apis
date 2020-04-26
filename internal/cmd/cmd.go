package cmd

import (
	_ "github.com/chirino/graphql-4-apis/internal/cmd/new"
	_ "github.com/chirino/graphql-4-apis/internal/cmd/serve"
	graphql_4_apis "github.com/chirino/graphql-4-apis/internal/cmd/root"
)

func Main() {
	graphql_4_apis.Main()
}
