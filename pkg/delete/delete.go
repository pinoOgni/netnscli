package delete

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pinoOgni/netnscli/pkg/testbed"
	vl "github.com/pinoOgni/netnscli/pkg/validator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configurationFilePath string
	// ErrDeleetLocalTestbed is returned when the deletion of the local testbed fails

	ErrDeleteLocalTestbed = fmt.Errorf("failed to delete local testbed")
)

// Cmd represents the delete command
var Cmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a local network testbed",
	Long:  `Starting from a yaml configuration file it deletes a local network testbed`,
	Run: func(cmd *cobra.Command, args []string) {
		testbed := testbed.FromFile(configurationFilePath)

		err := vl.ValidateConfiguration(testbed)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		log.Debugf("Unmarshalled testbed: %+v", testbed)

		// TODO is this needed?
		// Lock the OS thread to ensure namespace operations are consistent
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := testbed.Delete(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v", err)
			return
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&configurationFilePath, "file", "f", "", "path of the config file")
	if err := Cmd.MarkFlagRequired("file"); err != nil {
		panic("Configuration file is required")
	}
}
