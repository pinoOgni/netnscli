/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package create

import (
	"fmt"
	"runtime"

	"github.com/pinoOgni/netnscli/pkg/flags"
	"github.com/pinoOgni/netnscli/pkg/testbed"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configurationFilePath string

	// ErrCreateLocalTestbed is returned when the creation of the local testbed fails
	ErrCreateLocalTestbed = fmt.Errorf("failed to create local testbed")
)

// Cmd represents the create command
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a local network testbed",
	Long:  `Starting from a yaml configuration it creates a local network testbed.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Create command invoked")

		testbed := testbed.FromFile(configurationFilePath)
		log.Debugf("Unmarshalled testbed: %+v", testbed)

		// Lock the OS thread to ensure namespace operations are consistent
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// before create the namespaces, check if the user wants to delete the existing ones
		if cmd.Flags().Changed(flags.Force) {
			log.Debug("Deleting already existing namespaces before to apply the testbed")
			if err := testbed.DeleteNamespaces(); err != nil {
				log.Fatalf("could not delete existing namespaces: %v", err)
			}
		}

		if err := testbed.Apply(); err != nil {
			log.Fatalf("Could not apply the testbed: %v", err)
			return
		}

	},
}

func init() {
	Cmd.Flags().StringVarP(&configurationFilePath, "file", "f", "", "path of the config file")
	if err := Cmd.MarkFlagRequired("file"); err != nil {
		panic("Configuration file is required")
	}

	// Add the --force flag
	Cmd.Flags().Bool("force", false, "force the deletion of namespaces")
}
