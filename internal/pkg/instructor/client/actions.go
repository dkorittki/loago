package client

import (
	"context"

	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/rs/zerolog"
)

func Ping(ctx context.Context, logger *zerolog.Logger, client api.WorkerClient) error {
	res, err := client.Ping(ctx, &api.PingRequest{})

	if err != nil {
		return err
	}

	logger.Info().Str("response", res.Message).Msg("ping succeeded")

	return nil
}
