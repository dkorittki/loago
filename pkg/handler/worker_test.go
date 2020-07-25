package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dkorittki/loago-worker/pkg/service/loadtest"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"

	"github.com/dkorittki/loago-worker/pkg/api/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type RunServerMock struct {
	results []*api.EndpointResult
	mock.Mock
}

func (r *RunServerMock) Send(e *api.EndpointResult) error {
	r.results = append(r.results, e)
	args := r.Called(e)
	return args.Error(0)
}

func (r *RunServerMock) SetHeader(m metadata.MD) error {
	args := r.Called(m)
	return args.Error(0)
}

func (r *RunServerMock) SendHeader(m metadata.MD) error {
	args := r.Called(m)
	return args.Error(0)
}

func (r *RunServerMock) SetTrailer(m metadata.MD) {
	r.Called(m)
}

func (r *RunServerMock) Context() context.Context {
	args := r.Called()
	return args.Get(0).(context.Context)
}

func (r *RunServerMock) SendMsg(m interface{}) error {
	args := r.Called(m)
	return args.Error(0)
}

func (r *RunServerMock) RecvMsg(m interface{}) error {
	args := r.Called(m)
	return args.Error(0)
}

func TestNewWorker(t *testing.T) {
	w1 := &Worker{}
	w2 := NewWorker()

	assert.Equal(t, w1, w2)
}

func TestWorker_Run(t *testing.T) {
	srv := &RunServerMock{}
	srv.On("Send", mock.Anything).Return(nil).Once()
	srv.On("Send", mock.Anything).Return(nil).Once()
	srv.On("Send", mock.Anything).Return(status.Error(codes.Unavailable, "channel closed"))

	req := &api.RunRequest{
		Endpoints: []*api.RunRequest_Endpoint{
			{
				Url:    "http://foo.bar",
				Weight: 1,
			},
		},
		Amount:      1,
		Type:        api.RunRequest_FAKE,
		MinWaitTime: 1000,
		MaxWaitTime: 1000,
	}

	h := NewWorker()
	errChan := make(chan error)

	go func() {
		errChan <- h.Run(req, srv)
	}()

	err := <-errChan

	srv.AssertExpectations(t)
	assert.Equal(t, status.Error(codes.Unavailable, "channel closed"), err)
	assert.Len(t, srv.results, 3)

	for _, v := range srv.results {
		assert.Equal(t, int32(50), v.Ttfb)
		assert.Equal(t, int32(200), v.HttpStatusCode)
		assert.Equal(t, "OK", v.HttpStatusMessage)
		assert.False(t, v.Cached)
		assert.Equal(t, req.Endpoints[0].Url, v.Url)
	}
}

func TestServiceErrorHandling(t *testing.T) {
	srv := &RunServerMock{}
	req := &api.RunRequest{
		Endpoints: []*api.RunRequest_Endpoint{
			{
				Url:    "http://foo.bar",
				Weight: 1,
			},
		},
		Amount:      1,
		Type:        api.RunRequest_FAKE,
		MinWaitTime: 2000,
		MaxWaitTime: 1000,
	}

	h := NewWorker()
	errChan := make(chan error)

	go func() {
		errChan <- h.Run(req, srv)
	}()

	err := <-errChan

	assert.Equal(t, status.Error(codes.Aborted, loadtest.InvalidWaitBoundaries.Error()), err)
	assert.Len(t, srv.results, 0)
}

func TestInvalidBrowserType(t *testing.T) {
	srv := &RunServerMock{}
	req := &api.RunRequest{
		Endpoints: []*api.RunRequest_Endpoint{
			{
				Url:    "http://foo.bar",
				Weight: 1,
			},
		},
		Amount:      1,
		Type:        2,
		MinWaitTime: 2000,
		MaxWaitTime: 1000,
	}

	h := NewWorker()
	errChan := make(chan error)

	go func() {
		errChan <- h.Run(req, srv)
	}()

	err := <-errChan

	assert.Len(t, srv.results, 0)
	if assert.Error(t, err) {
		grpcErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, grpcErr.Code())
		assert.Contains(t, grpcErr.Message(), "unknown browser type")
	}
}

func TestUnkownErrorOnSend(t *testing.T) {
	srv := &RunServerMock{}
	srv.On("Send", mock.Anything).Return(errors.New("testing error"))

	req := &api.RunRequest{
		Endpoints: []*api.RunRequest_Endpoint{
			{
				Url:    "http://foo.bar",
				Weight: 1,
			},
		},
		Amount:      1,
		Type:        api.RunRequest_FAKE,
		MinWaitTime: 1000,
		MaxWaitTime: 1000,
	}

	h := NewWorker()
	errChan := make(chan error)

	go func() {
		errChan <- h.Run(req, srv)
	}()

	err := <-errChan

	assert.Len(t, srv.results, 1)
	if assert.Error(t, err) {
		grpcErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unknown, grpcErr.Code())
		assert.Contains(t, grpcErr.Message(), "no grpc error")
	}
}

func TestGRPCErrorOnSend(t *testing.T) {
	sendErr := status.Error(codes.Internal, "testing error")
	srv := &RunServerMock{}
	srv.On("Send", mock.Anything).Return(sendErr)

	req := &api.RunRequest{
		Endpoints: []*api.RunRequest_Endpoint{
			{
				Url:    "http://foo.bar",
				Weight: 1,
			},
		},
		Amount:      1,
		Type:        api.RunRequest_FAKE,
		MinWaitTime: 1000,
		MaxWaitTime: 1000,
	}

	h := NewWorker()
	errChan := make(chan error)

	go func() {
		errChan <- h.Run(req, srv)
	}()

	err := <-errChan
	assert.Len(t, srv.results, 1)
	assert.Equal(t, sendErr, err)
}
