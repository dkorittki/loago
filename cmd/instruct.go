package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dkorittki/loago/internal/pkg/instructor/client"
	"github.com/dkorittki/loago/pkg/instructor/config"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var instructor *client.Client

// instructCmd represents the instruct command
var instructCmd = &cobra.Command{
	Use:   "instruct",
	Short: "loago in instructor mode",
	Long: `Instructor mode is the client side of loago.
It manages worker instances, fetches loadtest results
and saves them.

Look at the subcommands 'ping' and 'run' for further
details.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		initClient()
	},
}

func init() {
	rootCmd.AddCommand(instructCmd)

	instructCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.loago.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tmp" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tmp")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file", viper.ConfigFileUsed())

		cfg, err := config.NewInstructorConfig(viper.Sub("instructor"))

		if err != nil {
			fmt.Println("Error parsing config file: ", err.Error())
			os.Exit(1)
		}

		instructorCfg = cfg
	} else {
		fmt.Println("no config file set, abort")
		os.Exit(1)
	}
}

func initClient() {
	instructor = client.NewClient()

	for _, v := range instructorCfg.Workers {
		certBytes, err := ioutil.ReadFile(v.Certificate)

		if err != nil {
			fmt.Println("cannot read certificate:", err.Error())
			os.Exit(1)
		}

		cert := tls.Certificate{Certificate: [][]byte{certBytes}}
		err = instructor.AddWorker(v.Adress, v.Port, v.Secret, &cert, nil)

		if err != nil {
			fmt.Println("cannot add worker to workerlist:", err.Error())
			os.Exit(1)
		}
	}
}
