package server

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/dkorittki/loago-worker/internal/pkg/handler"
	"github.com/dkorittki/loago-worker/pkg/api/v1"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcvalidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var authedMarker string

type Config struct {
	TlsCertPath  string
	TlsKeyPath   string
	Secret       string
	ListenAdress string
}

type WorkerServer struct {
	Server   *grpc.Server
	config   *Config
	listener net.Listener
}

func NewWorkerServer(cfg Config) (*WorkerServer, error) {
	lis, err := net.Listen("tcp", cfg.ListenAdress)
	if err != nil {
		return nil, err
	}

	return newWorkerServer(&cfg, handler.NewWorker(), lis)
}

func newWorkerServer(cfg *Config, handler api.WorkerServer, listener net.Listener) (*WorkerServer, error) {
	s := &WorkerServer{}
	s.config = cfg
	s.listener = listener

	var opts []grpc.ServerOption

	if cfg.TlsCertPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TlsCertPath, cfg.TlsKeyPath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}

	if cfg.Secret != "" {
		opts = append(opts, grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
			grpcauth.StreamServerInterceptor(authenticate(cfg.Secret)),
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

func (w *WorkerServer) Serve() error {
	return w.Server.Serve(w.listener)
}

func (w *WorkerServer) Stop() {
	w.Server.Stop()
}

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
