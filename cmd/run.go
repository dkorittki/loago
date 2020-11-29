package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dkorittki/loago/pkg/instructor/databackend"
	"github.com/dkorittki/loago/pkg/instructor/filedatabackend"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var (
	dataBackend databackend.DataBackend
	runCmd      = &cobra.Command{
		Use:      "run",
		Short:    "Run benchmarks",
		Long:     "Runs benchmarks on all configured workers and store the results on disk.",
		Run:      runRun,
		PreRunE:  preRunRun,
		PostRunE: postRunRun,
	}
)

func init() {
	instructCmd.AddCommand(runCmd)

	runCmd.Flags().String("result", "results.json", "Path to file in which the results will be stored")
}

func runRun(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info().Msg("Starting run request")

	results, err := instructor.Run(ctx, &logger, instructorCfg.Endpoints,
		instructorCfg.Amount, instructorCfg.MinWait, instructorCfg.MaxWait)

	if err != nil {
		logger.Error().Err(err).Msg("cannot initiate run request to workers")
		return
	}

	// stop requests on sigint and sigterm
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		logger.Debug().Msg("received sigint or sigterm, canceling requests")
		done <- true
	}()

	for {
		select {
		case res := <-results:
			logger.Debug().Msg("Received a result")
			if err := dataBackend.Store(&res); err != nil {
				logger.Warn().Err(err).Msg("Problem occurred while storing the result")
			}
		case <-done:
			logger.Info().Msg("Stopping requests to workers")
			return
		}
	}
}

func preRunRun(cmd *cobra.Command, args []string) error {
	logger.Info().Msg("Connecting to workers")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := instructor.Connect(ctx, &logger); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error().Err(err).Msg("timed out connection to workers")
		} else {
			logger.Error().Err(err).Msg("Cannot connect to all workers")
		}

		return err
	}

	logger.Info().Msg("Connections established!")
	logger.Info().Str("file", instructorCfg.ResultFile).Msg("Using file to store results")

	db, err := filedatabackend.New(instructorCfg.ResultFile)

	if err != nil {
		logger.Error().Err(err).Msg("Cannot open resultdata file")
		return err
	}

	dataBackend = db

	return nil
}

func postRunRun(cmd *cobra.Command, args []string) error {
	logger.Info().Msg("Disconnect from workers")

	err := instructor.Disconnect()

	if err != nil {
		logger.Error().Err(err).Msg("Cannot disconnect from all workers")
		return err
	}

	logger.Info().Msg("Connections closed!")
	logger.Info().Str("file", instructorCfg.ResultFile).Msg("Closing file")

	if dataBackend != nil {
		if err := dataBackend.Close(); err != nil {
			logger.Error().Err(err).Msg("Cannot close resultdata file")
		}
	}

	return nil
}
