package new

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/chirino/graphql-4-apis/internal/cmd/root"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "new",
		Short: "creates a new project with a default config",
		Run:   run,
		Args:  cobra.ExactArgs(1),
	}
	ConfigFile = ""
)

func init() {
	root.Command.AddCommand(Command)
}

func run(cmd *cobra.Command, args []string) {
	dir := args[0]
	os.MkdirAll(dir, 0755)

	configFile := filepath.Join(dir, "graphql-4-apis.yaml")
	err := ioutil.WriteFile(configFile, []byte(`#
# Configure the host and port the service will listen on
listen: 0.0.0.0:8080

# Configures how to get the openapi document.  It can be openapi v2 or v3.
Openapi:
  URL: openapi.json
  InsecureClient: false
  BearerToken:

# Configures the base URL that API requests will get issued against.
APIBase:
  URL: https://api.crc.testing:6443
  InsecureClient: true
  BearerToken:

`), 0644)

	if err != nil {
		log.Fatalf("%+v", err)
	}

	log.Printf(`Project created in the '%s' directory.`, dir)
	log.Printf(`Edit '%s' and then run:`, configFile)
	log.Println()
	log.Println(`    cd`, dir)
	log.Println(`    graphql-4-apis serve`)
	log.Println()
}
