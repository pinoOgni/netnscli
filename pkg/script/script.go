package script

import (
	"os"
	"text/template"

	"github.com/pinoOgni/netnscli/pkg/testbed"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configurationFilePath string
	output                string
)

const (
	defaultNs = "default"
)

// Cmd represents the create command
var Cmd = &cobra.Command{
	Use:   "script",
	Short: "Script creates a script from a yaml configuration file for a local network testbed",
	Long:  `Starting from a yaml configuration it creates the script to create it.`,
	Run: func(cmd *cobra.Command, args []string) {
		testbed := testbed.FromFile(configurationFilePath)
		log.Debugf("Unmarshalled testbed: %+v", testbed)

		// Load the template from file
		tmpl, err := template.New("script").Funcs(template.FuncMap{
			"isNotDefaultNamespace": func(namespace string) bool {
				return namespace != "" && namespace != defaultNs
			},
		}).Parse(scriptTemplate)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(output)
		if err != nil {
			log.Fatalf("Could not create %s: %s", output, err)
		}
		defer f.Close()

		// Execute the template and print to stdout
		err = tmpl.Execute(f, testbed)
		if err != nil {
			log.Errorf("Error executing template: %s", err)
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&configurationFilePath, "file", "f", "", "path of the config file")
	if err := Cmd.MarkFlagRequired("file"); err != nil {
		panic("Could not mark --file as a required flag")
	}
	Cmd.Flags().StringVarP(&output, "output", "o", "create_testbed.sh", "output script file (default is create_testbed.sh)")
}
