package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc/codes"

	"github.com/dkorittki/loago-worker/pkg/api/v1"
	"github.com/dkorittki/loago-worker/pkg/service/loadtest"
	loadtestservice "github.com/dkorittki/loago-worker/pkg/service/loadtest"
	"google.golang.org/grpc/status"
)

const ResultBufferSize = 1000

type Worker struct{}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Run(req *api.RunRequest, srv api.Worker_RunServer) error {
	ctx, cancel := context.WithCancel(context.Background())

	r := make(chan loadtest.EndpointResult, ResultBufferSize)
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
					return fmt.Errorf(errMsg+": %v", err)
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

func toServiceParams(req *api.RunRequest) (time.Duration, time.Duration, int,
	loadtest.BrowserType, []*loadtest.Endpoint, error) {
	minWait := time.Duration(req.MinWaitTime) * time.Millisecond
	maxWait := time.Duration(req.MaxWaitTime) * time.Millisecond
	amount := int(req.Amount)

	var browserType loadtest.BrowserType
	switch req.Type {
	case api.RunRequest_FAKE:
		browserType = loadtest.BrowserTypeFake
	case api.RunRequest_CHROME:
		browserType = loadtest.BrowserTypeChrome
	default:
		return 0, 0, 0, 0, nil,
			status.Errorf(codes.InvalidArgument, "invalid browser type identifier %d", req.Type)
	}

	var endpoints []*loadtest.Endpoint
	for _, v := range req.Endpoints {
		e := &loadtest.Endpoint{
			URL:    v.Url,
			Weight: uint(v.Weight),
		}
		endpoints = append(endpoints, e)
	}

	return minWait, maxWait, amount, browserType, endpoints, nil
}

func toRPCResponse(res *loadtest.EndpointResult) *api.EndpointResult {
	return &api.EndpointResult{
		Url:               res.URL,
		HttpStatusCode:    int32(res.HTTPStatusCode),
		HttpStatusMessage: res.HTTPStatusMessage,
		Ttfb:              int32(res.TTFB / time.Millisecond),
		Cached:            res.Cached,
	}
}
