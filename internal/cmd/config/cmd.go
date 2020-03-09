package api

import (
	"github.com/chirino/graphql-4-apis/internal/api"
	graphql_4_apis "github.com/chirino/graphql-4-apis/internal/cmd/graphql-4-apis"
	"github.com/chirino/graphql-4-apis/internal/gateway"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
)

var (
	Command = &cobra.Command{
		Use:   "config [config.yaml]",
		Short: "Runs the gateway using the specified config file",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
	ConfigFile = ""
)

func init() {
	_ = Command.MarkFlagRequired("config")
	graphql_4_apis.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	file, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatalf("%+v", err)
	}

	config := api.ApiResolverOptions{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	config.QueryType = `QueryApi`
	config.MutationType = `MutationApi`

	if !graphql_4_apis.Verbose {
		config.Logs = ioutil.Discard
	}
	gateway.ListenAndServe(config)
}
