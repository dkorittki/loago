package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "loago",
	Short: "A distributed loadtest utility for web apps",
	Long: `Loago is a distributed loadtest utility using Chromium based browsers
processes as HTTP clients.

Loago supports two modes:

worker mode: Perform the actual loadtest by request of a Loago instance in
instructor mode.

instructor mode: Configure and initiate a loadtest by distributing the loadtest
across one or more Loago instances in worker mode via network.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {}
