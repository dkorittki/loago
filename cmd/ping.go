package cmd

import (
	"context"
	"errors"
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
	Run:      runPing,
	PreRunE:  preRunPing,
	PostRunE: postRunPing,
}

func init() {
	instructCmd.AddCommand(pingCmd)
}

func runPing(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := instructor.Ping(ctx, &logger)

	if err != nil {
		e, ok := err.(client.WrappableError)

		if ok {
			logger.Error().Err(e.Unwrap()).Msg("cannot ping every worker")
		} else {
			logger.Error().Err(err).Msg("cannot ping every worker")
		}
	}
}

func preRunPing(cmd *cobra.Command, args []string) error {
	logger.Info().Msg("Connecting to workers")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := instructor.Connect(ctx, &logger)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error().Err(err).Msg("timed out connection to workers")
		} else {
			logger.Error().Err(err).Msg("Cannot connect to all workers")
		}

		return err
	}

	logger.Info().Msg("Connections established!")

	return nil
}

func postRunPing(cmd *cobra.Command, args []string) error {
	logger.Info().Msg("Disconnect from workers")

	err := instructor.Disconnect()

	if err != nil {
		logger.Error().Err(err).Msg("Cannot disconnect from all workers")
		return err
	}

	logger.Info().Msg("Connections closed!")

	return nil
}
