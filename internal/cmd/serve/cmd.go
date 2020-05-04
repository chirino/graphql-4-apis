package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/chirino/graphql-4-apis/internal/cmd/root"
	"github.com/chirino/graphql-4-apis/pkg/apis"
	"github.com/chirino/graphql/graphiql"
	"github.com/chirino/graphql/httpgql"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "serve",
		Short: "Runs the gateway service",
		Run:   run,
	}
	ConfigFile = ""
)

func init() {
	Command.Flags().StringVar(&ConfigFile, "config", "graphql-4-apis.yaml", "path to the config file to load")
	root.Command.AddCommand(Command)
}

type Config struct {
	Listen string `json:"listen"`
	apis.Config
}

func run(cmd *cobra.Command, args []string) {
	vebosityFmt := "%v"
	if !root.Verbose {
		vebosityFmt = "%+v\n"
	}

	file, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	config := Config{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	config.QueryType = `QueryApi`
	config.MutationType = `MutationApi`

	if !root.Verbose {
		config.Logs = ioutil.Discard
	}

	engine, err := apis.CreateGatewayEngine(config.Config)

	if config.Listen == "" {
		config.Listen = "0.0.0.0:8080"
	}

	host, port, err := net.SplitHostPort(config.Listen)
	if err != nil {
		log.Fatalf(vebosityFmt, err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port)
	http.Handle("/graphql", &httpgql.Handler{Engine: engine})
	log.Printf("GraphQL endpoint running at %s/graphql", endpoint)
	http.Handle("/", graphiql.New(endpoint+"/graphql", false))
	log.Printf("GraphQL UI running at %s", endpoint)

	log.Fatalf(vebosityFmt, http.ListenAndServe(config.Listen, nil))
}
