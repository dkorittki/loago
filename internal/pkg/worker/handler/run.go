package handler

import (
	"context"
	"time"

	loadtestservice "github.com/dkorittki/loago/internal/pkg/worker/service/loadtest"
	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Run handles incoming run requests. It starts a new loadtest
// and sends the response results of the runners
// as single messages via gRPC stream.
// Closing the gRPC channel stops the load test and shuts all runners down.
func (w *Worker) Run(req *api.RunRequest, srv api.Worker_RunServer) error {
	ctx, cancel := context.WithCancel(context.Background())

	r := make(chan loadtestservice.EndpointResult, ResultBufferSize)
	errChan := make(chan error)

	defer func() {
		cancel()
		close(r)
	}()

	minWait, maxWait, amount, browserType, endpoints, err := toServiceParams(req)
	if err != nil {
		return err
	}

	s := loadtestservice.New()
	go func() {
		errChan <- s.Run(ctx, browserType, endpoints, minWait, maxWait, amount, r)
	}()

	for {
		select {
		case err := <-errChan:
			close(errChan)
			return status.Error(codes.Aborted, err.Error())
		case res := <-r:
			err := srv.Send(toRPCResponse(&res))
			if err != nil {
				errStatus, ok := status.FromError(err)
				if !ok {
					errMsg := "received error which is no grpc error"
					log.Error().
						Str("component", "worker_handler").
						Err(err).
						Msg(errMsg)
					return status.Errorf(codes.Unknown, errMsg+"%v", err)
				}

				if errStatus.Code() == codes.Unavailable {
					log.Info().
						Str("component", "worker_handler").
						Msg("instructor closed connection")
				} else {
					log.Error().
						Str("component", "worker_handler").
						Err(err).
						Msg("unexpected error on transport")
				}

				return err
			}
		}
	}

}

// toServiceParams converts a gRPC API request data structure to seperate variables.
func toServiceParams(req *api.RunRequest) (time.Duration, time.Duration, int,
	loadtestservice.BrowserType, []*loadtestservice.Endpoint, error) {
	minWait := time.Duration(req.MinWaitTime) * time.Millisecond
	maxWait := time.Duration(req.MaxWaitTime) * time.Millisecond
	amount := int(req.Amount)

	var browserType loadtestservice.BrowserType
	switch req.Type {
	case api.RunRequest_FAKE:
		browserType = loadtestservice.BrowserTypeFake
	case api.RunRequest_CHROME:
		browserType = loadtestservice.BrowserTypeChrome
	default:
		return 0, 0, 0, 0, nil, ErrUnknownBrowser
	}

	var endpoints []*loadtestservice.Endpoint
	for _, v := range req.Endpoints {
		e := &loadtestservice.Endpoint{
			URL:    v.Url,
			Weight: uint(v.Weight),
		}
		endpoints = append(endpoints, e)
	}

	return minWait, maxWait, amount, browserType, endpoints, nil
}

// toRPCResponse converts a service endpoint result data structure to an gRPC API endpointresult
func toRPCResponse(res *loadtestservice.EndpointResult) *api.EndpointResult {
	return &api.EndpointResult{
		Url:               res.URL,
		HttpStatusCode:    int32(res.HTTPStatusCode),
		HttpStatusMessage: res.HTTPStatusMessage,
		Ttfb:              int32(res.TTFB / time.Millisecond),
		Cached:            res.Cached,
	}
}
