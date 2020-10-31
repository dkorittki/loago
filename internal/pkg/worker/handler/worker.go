// Package handler provides request handler implementations.
package handler

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ResultBufferSize sets the amount of objects in the result buffer.
const ResultBufferSize = 1000

var (
	// ErrUnknownBrowser indicates an error when an unknown browser type is given.
	ErrUnknownBrowser = status.Error(codes.InvalidArgument, "unknown browser type in request")
)

// Worker implements the gRPC worker service handler.
type Worker struct{}

// NewWorker returns a new Worker.
func NewWorker() *Worker {
	return &Worker{}
}
