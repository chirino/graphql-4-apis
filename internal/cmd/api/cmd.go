package api

import (
	"github.com/chirino/graphql-4-apis/internal/api"
	graphql_4_apis "github.com/chirino/graphql-4-apis/internal/cmd/graphql-4-apis"
	"github.com/chirino/graphql-4-apis/internal/gateway"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	Command = &cobra.Command{
		Use:   "api [openapi url or file path]",
		Short: "Runs the gateway using the specified openapi document",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	Insecure = false
)

func init() {
	Command.Flags().BoolVar(&Insecure, "insecure", false, "accept invalid https server certificates")
	graphql_4_apis.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	config := api.ApiResolverOptions{}
	config.Openapi.URL = args[0]
	config.Openapi.InsecureClient = Insecure
	config.APIBase.InsecureClient = Insecure
	config.QueryType = `QueryApi`
	config.MutationType = `MutationApi`
	if !graphql_4_apis.Verbose {
		config.Logs = ioutil.Discard
	}
	gateway.ListenAndServe(config)

}
