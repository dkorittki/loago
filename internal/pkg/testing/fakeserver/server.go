package fakeserver

import (
	"context"

	"github.com/dkorittki/loago/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FakeWorkerServer struct{}

func (s *FakeWorkerServer) Ping(_ context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	res := &api.PingResponse{
		Message: "test",
	}

	return res, nil
}

func (s *FakeWorkerServer) Run(req *api.RunRequest, srv api.Worker_RunServer) error {
	return status.Errorf(codes.Unimplemented, "method Run not implemented")
}
