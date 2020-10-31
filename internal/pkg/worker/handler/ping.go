package handler

import (
	"context"
	"errors"

	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/peer"
)

// Ping responds with a PingResponse message to incoming ping requests.
func (w *Worker) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	p, ok := peer.FromContext(ctx)

	if !ok {
		err := errors.New("can't determine peer information from request")
		log.Error().Msg(err.Error())
		return nil, err
	}

	log.Info().Str("source ip", p.Addr.String()).Msg("incoming ping request")

	return &api.PingResponse{Message: "pong"}, nil
}
