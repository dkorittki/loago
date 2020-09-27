// Package server provides a server for handling worker gRPC communication.
package server

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/dkorittki/loago/internal/pkg/handler"
	"github.com/dkorittki/loago/pkg/api/v1"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcvalidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var authedMarker struct{}

// Config configures a server.
type Config struct {
	// Path to TLS certificate used for encryption on gRPC connections.
	// If empty, no TLS connection is used.
	TLSCertPath string

	// Path to TLS private key used for encryption on gRPC connections.
	// Only use this in conjunction with TLSCertPath.
	TLSKeyPath string

	// Secret is a basic auth token used for authentication.
	// Only use this in conjunction with TLSCertPath and TLSKeyPath.
	Secret string

	// ListenAdress contains the interface ip and port to listen on,
	// i.e. "127.0.0.1:50051".
	ListenAdress string
}

// WorkerServer is a server for handling worker gRPC requests.
// It implements the Worker interface.
type WorkerServer struct {
	// Server is the gRPC server.
	Server *grpc.Server

	// Listener contains the network connection listener.
	listener net.Listener
}

// NewWorkerServer returns a new WorkerServer or nil and an error,
// when something went wrong.
// It tries to create a network listener, loads TLS certificates and keys and
// configures authentication and request validation.
func NewWorkerServer(cfg Config) (*WorkerServer, error) {
	lis, err := net.Listen("tcp", cfg.ListenAdress)
	if err != nil {
		return nil, err
	}

	var cert tls.Certificate
	if cfg.TLSCertPath != "" {
		cert, err = tls.LoadX509KeyPair(cfg.TLSCertPath, cfg.TLSKeyPath)
		if err != nil {
			return nil, err
		}
	}

	return newWorkerServer(&cert, cfg.Secret, handler.NewWorker(), lis)
}

func newWorkerServer(cert *tls.Certificate, secret string, handler api.WorkerServer,
	listener net.Listener) (*WorkerServer, error) {
	s := &WorkerServer{}
	s.listener = listener

	var opts []grpc.ServerOption

	if cert != nil && len(cert.Certificate) != 0 {
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(cert)))
	}

	if secret != "" {
		opts = append(opts, grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
			grpcauth.StreamServerInterceptor(authenticate(secret)),
			grpcvalidator.StreamServerInterceptor(),
		)))
	} else {
		opts = append(opts, grpc.StreamInterceptor(
			grpcvalidator.StreamServerInterceptor(),
		))
	}

	server := grpc.NewServer(opts...)
	api.RegisterWorkerServer(server, handler)
	s.Server = server

	return s, nil
}

// Serve starts the gRPC server, listening for incoming requests.
func (w *WorkerServer) Serve() error {
	return w.Server.Serve(w.listener)
}

// Stop stops the gRPC server, not listening for incoming requests anymore.
func (w *WorkerServer) Stop() {
	w.Server.Stop()
}

// authenticate returns a new function checking the correctness of a secret token.
func authenticate(secret string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpcauth.AuthFromMD(ctx, "basic")
		if err != nil {
			return nil, err
		}

		if token != secret {
			return nil, status.Errorf(codes.PermissionDenied, "wrong auth secret")
		}

		return context.WithValue(ctx, authedMarker, "exists"), nil
	}
}
