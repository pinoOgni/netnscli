/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package create

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/pinoOgni/netnscli/pkg/flags"
	"github.com/pinoOgni/netnscli/pkg/netlink"
	"github.com/pinoOgni/netnscli/pkg/netns"
	"github.com/pinoOgni/netnscli/pkg/testbed"
	vl "github.com/pinoOgni/netnscli/pkg/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	netnsv "github.com/vishvananda/netns"
)

var (
	cfgFile string
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
				_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}

			// Unmarshal the config into the Config struct
			err = viper.Unmarshal(&testbed)
			if err != nil {
				log.Fatalf("Unable to decode into struct %v", err)
			}
			err = vl.ValidateConfiguration(testbed)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}
			// Print the Config struct to see if the data is loaded correctly
			fmt.Printf("Testbed: %+v\n", testbed)
			// Get all settings as a map
			// settings := viper.AllSettings()

			// Print the settings
			// fmt.Printf("Config values: \n")
			// for key, value := range settings {
			// fmt.Printf("%s : %v\n", key, value)
			// }

			// create namespaces
			// Lock the OS thread to ensure namespace operations are consistent
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			// Save the current network namespace because the CreateNamespace will move
			// the context on the created namespace

			// TODO actually I don't think that this is needed.. only the defer close
			// in the CreateNamespace should be the correct way
			origNS, err := netnsv.Get()
			if err != nil {
				log.Fatalf("Failed to get current network namespace: %v", err)
			}
			defer origNS.Close()
			// loop to delete all existing namespaces and then create new ones
			// TODO check if anothe logic can be implemented and chosen by the user in some way

			// TODO aldo comment --force to delete the existing namespace
			// default does not delete it
			for _, nsName := range testbed.Namespaces {
				if err := netns.DeleteNamespace(nsName.Name); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				if err := netns.CreateNamespace(nsName.Name); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}
			// Switch back to the original namespace
			err = netnsv.Set(origNS)
			if err != nil {
				log.Fatalf("Failed to switch back to the original network namespace: %v", err)
			}

			// create veth pairs
			for _, vethPair := range testbed.VethPairs {
				if err := netlink.CreateVethPair(vethPair); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}

			// set veths ns
			for _, vethPair := range testbed.VethPairs {
				if err := netlink.SetVethPairNs(vethPair); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}
			// create bridges
			for _, bridge := range testbed.Bridges {
				if err := netlink.CreateBridge(bridge); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}

			// set bridges
			for _, bridge := range testbed.Bridges {
				if err := netlink.SetUpAndAttachInterfacesToBridge(bridge); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}

			// set up veths
			for _, vethPair := range testbed.VethPairs {
				if err := netlink.SetVethPairUp(vethPair); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}

			// add address to veths
			for _, vethPair := range testbed.VethPairs {
				if err := netlink.AddAddressVethPair(vethPair); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Error: flag --file must be used. \n"+
				"See 'netnscli create --help' for help and examples.\n")
			return
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	Cmd.Flags().StringVarP(&cfgFile, "file", "f", "", "config file is required")
	if err := viper.BindPFlag("file", Cmd.Flag("file")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: unable to bind flag file %v\n", err)
	}
	// Bind all persistent flags to viper
	if err := viper.BindPFlags(Cmd.PersistentFlags()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
