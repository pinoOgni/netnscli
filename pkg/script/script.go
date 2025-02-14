package script

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/pinoOgni/netnscli/pkg/flags"
	"github.com/pinoOgni/netnscli/pkg/testbed"
	vl "github.com/pinoOgni/netnscli/pkg/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	output  string

	// ErrScriptTestbed is returned when the creation of script for the local testbed fails
	ErrScriptTestbed = fmt.Errorf("failed to create script for local testbed")
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

			if err := createScriptFile(&testbed); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}

		} else {
			_, _ = fmt.Fprintf(os.Stderr, "error: flag --file must be used. \n"+
				"See 'netnscli script --help' for help and examples.\n")
			return
		}
	},
}

// func createScriptFile(testbed *testbed.Configuration) error {
// 	// Create the script file
// 	script, err := os.Create(output)
// 	if err != nil {
// 		log.Fatalf("error: unable to create script file: %v", err)
// 	}
// 	defer script.Close()

// 	// Write the script header
// 	_, _ = script.WriteString("#!/bin/bash\n\n")

// 	_, _ = script.WriteString("# add namespaces\n")
// 	for _, nsName := range testbed.Namespaces {
// 		command := fmt.Sprintf("ip netns add %s\n", nsName.Name)
// 		_, _ = script.WriteString(command)
// 	}

// 	// create veth pairs and set veths ns
// 	_, _ = script.WriteString("\n# create veth pairs and set it to the correct ns\n")
// 	for _, vethPair := range testbed.VethPairs {
// 		command := fmt.Sprintf("ip link add %s type veth peer name %s \n", vethPair.P1Name, vethPair.P2Name)
// 		_, _ = script.WriteString(command)
// 		if vethPair.P1Namespace != defaultNs {
// 			command = fmt.Sprintf("ip link set %s netns %s \n", vethPair.P1Name, vethPair.P1Namespace)
// 			_, _ = script.WriteString(command)
// 		}
// 		if vethPair.P2Namespace != defaultNs {
// 			command = fmt.Sprintf("ip link set %s netns %s \n", vethPair.P2Name, vethPair.P2Namespace)
// 			_, _ = script.WriteString(command)
// 		}
// 		_, _ = script.WriteString("\n")
// 	}

// 	// create bridges
// 	_, _ = script.WriteString("\n# create the bridge in the default namespace \n")
// 	for _, bridge := range testbed.Bridges {
// 		command := fmt.Sprintf("ip link add name %s type bridge\n", bridge.Name)
// 		_, _ = script.WriteString(command)
// 	}

// 	// set bridges (set it up and attach interfaces)
// 	_, _ = script.WriteString("\n# set bridge up and attach veth interfaces to the bridge \n")
// 	for _, bridge := range testbed.Bridges {
// 		command := fmt.Sprintf("ip link set %s up\n", bridge.Name)
// 		_, _ = script.WriteString(command)
// 		for _, iface := range bridge.Interfaces {
// 			command = fmt.Sprintf("ip link set %s master %s\n", iface, bridge.Name)
// 			_, _ = script.WriteString(command)
// 		}
// 	}

// 	// bring up the veth interfaces in the default namespace
// 	_, _ = script.WriteString("\n# bring up the veth interfaces in the default namespace \n")
// 	for _, vethPair := range testbed.VethPairs {
// 		if vethPair.P1Namespace == defaultNs {
// 			command := fmt.Sprintf("ip link set %s up\n", vethPair.P1Name)
// 			_, _ = script.WriteString(command)
// 		}
// 		if vethPair.P2Namespace == defaultNs {
// 			command := fmt.Sprintf("ip link set %s up\n", vethPair.P2Name)
// 			_, _ = script.WriteString(command)
// 		}
// 	}

// 	// add address to veths in default and network namespaces and set it up
// 	for _, vethPair := range testbed.VethPairs {
// 		if vethPair.P1Namespace != defaultNs {
// 			command := fmt.Sprintf("ip netns exec %s ip addr add %s dev %s \n", vethPair.P1Namespace, vethPair.P1IPAddress, vethPair.P1Name)
// 			_, _ = script.WriteString(command)
// 			command = fmt.Sprintf("ip netns exec %s ip link set %s up \n", vethPair.P1Namespace, vethPair.Name)
// 			_, _ = script.WriteString(command)
// 		} else {
// 			command := fmt.Sprintf("ip addr add %s dev %s \n", vethPair.P1IPAddress, vethPair.P1Name)
// 			_, _ = script.WriteString(command)
// 			command = fmt.Sprintf("ip link set %s up \n", vethPair.P1Name)
// 			_, _ = script.WriteString(command)
// 		}
// 		if vethPair.P2Namespace != defaultNs {
// 			command := fmt.Sprintf("ip netns exec %s ip addr add %s dev %s \n", vethPair.P2Namespace, vethPair.P2IPAddress, vethPair.P2Name)
// 			_, _ = script.WriteString(command)
// 			command = fmt.Sprintf("ip netns exec %s ip link set %s up \n", vethPair.P1Namespace, vethPair.P2Name)
// 			_, _ = script.WriteString(command)
// 		} else {
// 			command := fmt.Sprintf("ip addr add %s dev %s \n", vethPair.P2IPAddress, vethPair.P2Name)
// 			_, _ = script.WriteString(command)
// 			command = fmt.Sprintf("ip link set %s up \n", vethPair.P2Name)
// 			_, _ = script.WriteString(command)
// 		}
// 	}
// 	return nil
// }

func createScriptFile(testbed *testbed.Configuration) error {
	// Create the script file
	script, err := os.Create(output)
	if err != nil {
		log.Fatalf("error: unable to create script file: %v", err)
	}
	defer script.Close()

	var buffer bytes.Buffer

	// Write the script header
	buffer.WriteString("#!/bin/bash\n\n")

	// Add namespaces
	buffer.WriteString("# add namespaces\n")
	for _, nsName := range testbed.Namespaces {
		buffer.WriteString(fmt.Sprintf("ip netns add %s\n", nsName.Name))
	}

	// Create veth pairs and set them in namespaces
	buffer.WriteString("\n# create veth pairs and set them to the correct ns\n")
	for _, vethPair := range testbed.VethPairs {
		buffer.WriteString(fmt.Sprintf("ip link add %s type veth peer name %s\n", vethPair.P1.Name, vethPair.P2.Name))
		if vethPair.P1.Namespace != defaultNs {
			buffer.WriteString(fmt.Sprintf("ip link set %s netns %s\n", vethPair.P1.Name, vethPair.P1.Namespace))
		}
		if vethPair.P2.Namespace != defaultNs {
			buffer.WriteString(fmt.Sprintf("ip link set %s netns %s\n", vethPair.P2.Name, vethPair.P2.Namespace))
		}
		buffer.WriteString("\n")
	}

	// Create bridges
	buffer.WriteString("\n# create the bridge in the default namespace\n")
	for _, bridge := range testbed.Bridges {
		buffer.WriteString(fmt.Sprintf("ip link add name %s type bridge\n", bridge.Name))
	}

	// Set bridges up and attach interfaces
	buffer.WriteString("\n# set bridges up and attach veth interfaces to the bridge\n")
	for _, bridge := range testbed.Bridges {
		buffer.WriteString(fmt.Sprintf("ip link set %s up\n", bridge.Name))
		for _, iface := range bridge.Interfaces {
			buffer.WriteString(fmt.Sprintf("ip link set %s master %s\n", iface, bridge.Name))
		}
		buffer.WriteString("\n")
	}

	// Bring up the veth interfaces in the default namespace
	buffer.WriteString("\n# bring up the veth interfaces in the default namespace\n")
	for _, vethPair := range testbed.VethPairs {
		if vethPair.P1.Namespace == defaultNs {
			buffer.WriteString(fmt.Sprintf("ip link set %s up\n", vethPair.P1.Name))
		}
		if vethPair.P2.Namespace == defaultNs {
			buffer.WriteString(fmt.Sprintf("ip link set %s up\n", vethPair.P2.Name))
		}
	}

	// Add addresses to veths and set them up
	buffer.WriteString("\n# add addresses to veths and set them up\n")
	for _, vethPair := range testbed.VethPairs {
		if vethPair.P1.Address != "" {
			if vethPair.P1.Namespace != defaultNs {
				buffer.WriteString(fmt.Sprintf("ip netns exec %s ip addr add %s dev %s\n", vethPair.P1.Namespace, vethPair.P1.Address, vethPair.P1.Name))
				buffer.WriteString(fmt.Sprintf("ip netns exec %s ip link set %s up\n", vethPair.P1.Namespace, vethPair.P1.Name))
			} else {
				buffer.WriteString(fmt.Sprintf("ip addr add %s dev %s\n", vethPair.P1.Address, vethPair.P1.Name))
				buffer.WriteString(fmt.Sprintf("ip link set %s up\n", vethPair.P1.Name))
			}
			buffer.WriteString("\n")
		}
		if vethPair.P2.Address != "" {
			if vethPair.P2.Namespace != defaultNs {
				buffer.WriteString(fmt.Sprintf("ip netns exec %s ip addr add %s dev %s\n", vethPair.P2.Namespace, vethPair.P2.Address, vethPair.P2.Name))
				buffer.WriteString(fmt.Sprintf("ip netns exec %s ip link set %s up\n", vethPair.P2.Namespace, vethPair.P2.Name))
			} else {
				buffer.WriteString(fmt.Sprintf("ip addr add %s dev %s\n", vethPair.P2.Address, vethPair.P2.Name))
				buffer.WriteString(fmt.Sprintf("ip link set %s up\n", vethPair.P2.Name))
			}
			buffer.WriteString("\n")
		}
	}

	// Write all commands to the script file at once
	_, err = script.WriteString(buffer.String())
	return err
}

func init() {
	cobra.OnInitialize(initConfig)
	Cmd.Flags().StringVarP(&cfgFile, "file", "f", "", "config file is required")
	if err := viper.BindPFlag("file", Cmd.Flag("file")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: unable to bind flag file %v\n", err)
	}
	// Add the --output flag with a default value
	Cmd.Flags().StringVarP(&output, "output", "o", "create_testbed.sh", "output script file (default is create_testbed.sh)")
	if err := viper.BindPFlag("output", Cmd.Flag("output")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: unable to bind flag output %v\n", err)
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
