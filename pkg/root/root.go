package root

import (
	"os"

	"github.com/pinoOgni/netnscli/pkg/create"
	"github.com/pinoOgni/netnscli/pkg/delete"
	"github.com/pinoOgni/netnscli/pkg/model"
	"github.com/pinoOgni/netnscli/pkg/script"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netns"
)

const (
	logo = `
  _  _ ___ _____ _  _ ___  ___ _    ___ 
 | \| | __|_   _| \| / __|/ __| |  |_ _|
 | .  | _|  | | |  ' \__ \ (__| |__ | | 
 |_|\_|___| |_| |_|\_|___/\___|____|___|
                                        
netnscli creates and manages local network testbed			   
`
)

var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "netnscli",
	Short: "netnscli creates a local network testbed",
	Long:  logo,
	Run:   nil,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(create.Cmd)
	rootCmd.AddCommand(delete.Cmd)
	rootCmd.AddCommand(script.Cmd)

	rootCmd.PersistentFlags().BoolVar(&verbose, "debug", false, "Show a more verbose output logs")

	cobra.OnInitialize(initLogger, initProgramNamespace)
}

func initLogger() {
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func initProgramNamespace() {
	namespaceHandle, err := netns.Get()
	if err != nil {
		panic("could not get the current namespace")
	}
	model.ProgramNamespace = namespaceHandle
}
