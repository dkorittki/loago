package cmd

import (
	"context"
	"time"

	"github.com/dkorittki/loago/internal/pkg/instructor/client"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Test connection to workers",
	Long: `Ping tests the connectivity to workers specific in --config.
If something is off with the TLS configuration, the authentication secret,
or in case of network problems, this command will help.`,
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
	instructCmd.AddCommand(pingCmd)
}
