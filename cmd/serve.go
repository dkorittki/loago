package cmd

import (
	"fmt"

	"github.com/dkorittki/loago/internal/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	addr     string
	port     int
	secret   string
	certPath string
	keyPath  string

	// serveCmd represents the serve command
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Act as a worker",
		Long: `Start loago in worker mode, which performs loadtests.

Loago will bind on an configurable interface/port and waits for requests
coming from other loago instances in instructor mode. It will use a Chrome
based browser to perform the loadtest.

Make sure you have Chrome or Chromium installed on the system, where you want
to use Loago in worker mode.`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := server.Config{
				TLSCertPath:  certPath,
				TLSKeyPath:   keyPath,
				Secret:       secret,
				ListenAdress: fmt.Sprintf("%s:%d", addr, port),
			}

			log.Info().Str("listen_adress", cfg.ListenAdress).Msg("start serving")

			ws, err := server.NewWorkerServer(cfg)
			if err != nil {
				log.Fatal().Err(err).Msg("error on starting server")
			}

			if err := ws.Serve(); err != nil {
				log.Fatal().Err(err).Msg("error on serving")
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&addr, "adress", "0.0.0.0", "listen address, e.g. '127.0.0.1' or '0.0.0.0'")
	serveCmd.Flags().IntVar(&port, "port", 50051, "listen port")
	serveCmd.Flags().StringVar(&secret, "secret", "", "basic auth secret used between a worker and an instructor")
	serveCmd.Flags().StringVar(&certPath, "cert", "", "path to TLS certificate")
	serveCmd.Flags().StringVar(&keyPath, "key", "", "path to TLS key")
}
