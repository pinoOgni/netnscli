package delete

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/pinoOgni/netnscli/internal/netlink"
	"github.com/pinoOgni/netnscli/internal/netns"
	"github.com/pinoOgni/netnscli/pkg/flags"
	"github.com/pinoOgni/netnscli/pkg/testbed"
	vl "github.com/pinoOgni/netnscli/pkg/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// ErrDeleetLocalTestbed is returned when the deletion of the local testbed fails

	ErrDeleteLocalTestbed = fmt.Errorf("failed to delete local testbed")
)

// Cmd represents the delete command
var Cmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a local network testbed",
	Long:  `Starting from a yaml configuration file it deletes a local network testbed`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO move this part and the one in the create command in a pkg
		var testbed testbed.Configuration
		if cmd.Flags().Changed(flags.File) {
			err := viper.ReadInConfig()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "error: flag --file must be used. \n"+
				"See 'netnscli create --help' for help and examples.\n")
			return
		}
		// Unmarshal the config into the Config struct
		err := viper.Unmarshal(&testbed)
		if err != nil {
			log.Fatalf("Unable to decode into struct %v", err)
		}
		err = vl.ValidateConfiguration(testbed)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		// Print the Config struct to see if the data is loaded correctly
		//fmt.Printf("Testbed: %+v\n", testbed)

		// TODO is this needed?
		// Lock the OS thread to ensure namespace operations are consistent
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := deleteLocalTestbed(&testbed); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
	},
}

func deleteLocalTestbed(testbed *testbed.Configuration) error {
	// delete all namespaces
	for _, nsName := range testbed.Namespaces {
		if err := netns.DeleteNamespace(nsName.Name); err != nil {
			return fmt.Errorf("%w: %v", ErrDeleteLocalTestbed, err)
		}
	}
	// detach all interfaces from bridges
	for _, bridge := range testbed.Bridges {
		if err := netlink.DetachAllInterfacesFromBridge(bridge); err != nil {
			return fmt.Errorf("%w: %v", ErrDeleteLocalTestbed, err)

		}
	}
	// delete bridges
	for _, bridge := range testbed.Bridges {
		if err := netlink.DeleteBridge(bridge); err != nil {
			return fmt.Errorf("%w: %v", ErrDeleteLocalTestbed, err)
		}
	}
	return nil
}

func init() {

	cobra.OnInitialize(initConfig)
	Cmd.Flags().StringVarP(&cfgFile, "file", "f", "", "config file is required")
	if err := viper.BindPFlag("file", Cmd.Flag("file")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: unable to bind flag file %v\n", err)
	}
	// Bind all persistent flags to viper
	if err := viper.BindPFlags(Cmd.PersistentFlags()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func initConfig() {
	// set default parameters
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	// If a configuration file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		return
	}
}
