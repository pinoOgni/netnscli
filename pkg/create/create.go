/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package create

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
	netnsv "github.com/vishvananda/netns"
)

var (
	cfgFile string

	// ErrCreateLocalTestbed is returned when the creation of the local testbed fails
	ErrCreateLocalTestbed = fmt.Errorf("failed to create local testbed")
)

// Cmd represents the create command
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a local network testbed",
	Long:  `Starting from a yaml configuration it creates a local network testbed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var testbed testbed.Configuration
		if cmd.Flags().Changed(flags.File) {
			err := viper.ReadInConfig()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}

			// Unmarshal the config into the testbed configuration
			err = viper.Unmarshal(&testbed)
			if err != nil {
				log.Fatalf("Unable to decode into struct %v", err)
			}
			err = vl.ValidateConfiguration(testbed)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}
			// TODO debug: print the Config struct to see if the data is loaded correctly
			//fmt.Printf("Testbed: %+v\n", testbed)

			// Note: leave it for debug purpose
			// Get all settings as a map
			// settings := viper.AllSettings()

			// Print the settings
			// fmt.Printf("Config values: \n")
			// for key, value := range settings {
			// fmt.Printf("%s : %v\n", key, value)
			// }

			// Lock the OS thread to ensure namespace operations are consistent
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			// Save the current network namespace because the addition of a new
			// network namespace will move the context on the added network namespace

			// TODO actually I don't think that this is needed.. only the defer close
			// in the Add should be the correct way
			originNs, err := netnsv.Get()
			if err != nil {
				log.Fatalf("Failed to get current network namespace: %v", err)
			}
			defer originNs.Close()

			// before create the namespaces, check if the user wants to delete the existing ones
			if cmd.Flags().Changed(flags.Force) {
				if err := deleteNamespaces(&testbed); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
					return
				}
			}
			if err := createLocalTestbed(originNs, &testbed); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}

		} else {
			_, _ = fmt.Fprintf(os.Stderr, "error: flag --file must be used. \n"+
				"See 'netnscli create --help' for help and examples.\n")
			return
		}
	},
}

func deleteNamespaces(testbed *testbed.Configuration) error {
	for _, nsName := range testbed.Namespaces {
		if err := netns.DeleteNamespace(nsName.Name); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}
	return nil
}
func createLocalTestbed(originNs netnsv.NsHandle, testbed *testbed.Configuration) error {
	// TODO --force to delete the existing namespace, the default behavior does not delete it
	for _, nsName := range testbed.Namespaces {
		if err := netns.Add(nsName.Name); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}
	// Switch back to the original namespace
	err := netnsv.Set(originNs)
	if err != nil {
		log.Fatalf("Failed to switch back to the original network namespace: %v", err)
	}

	// create veth pairs
	for _, vethPair := range testbed.VethPairs {
		if err := netlink.CreateVethPair(vethPair); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}

	// set veths ns
	for _, vethPair := range testbed.VethPairs {
		if err := netlink.SetVethPairNs(vethPair); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}
	// create bridges
	for _, bridge := range testbed.Bridges {
		if err := netlink.CreateBridge(bridge); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}

	// set bridges
	for _, bridge := range testbed.Bridges {
		if err := netlink.SetUpAndAttachInterfacesToBridge(bridge); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}

	// set up veths
	for _, vethPair := range testbed.VethPairs {
		if err := netlink.SetVethPairUp(vethPair); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
		}
	}

	// add address to veths
	for _, vethPair := range testbed.VethPairs {
		if err := netlink.AddAddressVethPair(vethPair); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateLocalTestbed, err)
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
	// Add the --force flag
	Cmd.Flags().Bool("force", false, "force the deletion of namespaces")
	if err := viper.BindPFlag("force", Cmd.Flag("force")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: unable to bind flag force %v\n", err)
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
