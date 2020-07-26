package server

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/dkorittki/loago-worker/internal/pkg/handler"
	"github.com/dkorittki/loago-worker/pkg/api/v1"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
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
	Server *grpc.Server
	config *Config
}

func NewWorkerServer(cfg Config) (*WorkerServer, error) {
	return newWorkerServer(&cfg, handler.NewWorker())
}

func newWorkerServer(cfg *Config, handler api.WorkerServer) (*WorkerServer, error) {
	s := &WorkerServer{}
	s.config = cfg

	var opts []grpc.ServerOption

	if cfg.TlsCertPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TlsCertPath, cfg.TlsKeyPath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}

	if cfg.Secret != "" {
		opts = append(opts, grpc.StreamInterceptor(grpcauth.StreamServerInterceptor(authenticate(cfg.Secret))))
	}

	server := grpc.NewServer(opts...)
	api.RegisterWorkerServer(server, handler)
	s.Server = server

	return s, nil
}

func (w *WorkerServer) Serve() error {
	lis, err := net.Listen("tcp", w.config.ListenAdress)
	if err != nil {
		return err
	}

	return w.Server.Serve(lis)
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
