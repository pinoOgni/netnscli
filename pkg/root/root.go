package root

import (
	"fmt"
	"os"

	"github.com/pinoOgni/netnscli/pkg/create"
	"github.com/pinoOgni/netnscli/pkg/delete"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "netnscli",
	Short: "netnscli creates a local network testbed",
	Long:  logo,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
	},
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
	cmdFlags := rootCmd.PersistentFlags()
	err := viper.BindPFlags(cmdFlags)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
}
