package main

import (
	"flag"
	"fmt"

	"github.com/dkorittki/loago-worker/internal/pkg/server"
	"github.com/rs/zerolog/log"
)

func main() {
	addr := flag.String("adress", "0.0.0.0", "listen address, e.g. '127.0.0.1' or '0.0.0.0'")
	port := flag.Int("port", 50051, "listen port")
	secret := flag.String("secret", "", "basic auth secret, which the instructor must use")
	certPath := flag.String("cert", "", "path to TLS certificate")
	keyPath := flag.String("key", "", "path to TLS key")
	flag.Parse()

	cfg := server.Config{
		TlsCertPath:  *certPath,
		TlsKeyPath:   *keyPath,
		Secret:       *secret,
		ListenAdress: fmt.Sprintf("%s:%d", *addr, *port),
	}

	log.Info().Str("listen_adress", cfg.ListenAdress).Msg("start serving")

	ws, err := server.NewWorkerServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error on starting server")
	}

	if err := ws.Serve(); err != nil {
		log.Fatal().Err(err).Msg("error on serving")
	}
}
