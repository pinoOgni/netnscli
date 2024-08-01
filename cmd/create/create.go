/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package create

import (
	"fmt"
	"log"
	"os"

	"github.com/pinoOgni/netnscli/pkg/flags"
	"github.com/pinoOgni/netnscli/pkg/netlink"
	"github.com/pinoOgni/netnscli/pkg/netns"
	"github.com/pinoOgni/netnscli/pkg/testbed"
	vl "github.com/pinoOgni/netnscli/pkg/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Error: flag --file must be used. \n"+
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
		for _, nsName := range testbed.Namespaces {
			if err := netns.CreateNamespace(nsName.Name); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}

		// create and set up veths
		if err := netlink.CreateVethPairs(testbed.VethPairs); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		// create bridge

		// configure veths
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
