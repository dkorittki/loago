package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/dkorittki/loago/internal/pkg/instructor/client"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empokjnbwers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := instructor.ExecuteAction(ctx, "ping", &logger)

		if err != nil {
			e, ok := err.(client.WrappableError)

			if ok {
				logger.Error().Err(e.Unwrap()).Msg("cannot ping every worker")
			} else {
				logger.Error().Err(err).Msg("cannot ping every worker")
			}
		}
	},
}

func init() {
	fmt.Println("init ping called")
	instructCmd.AddCommand(pingCmd)
}
