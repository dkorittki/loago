package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/test/bufconn"

	"github.com/dkorittki/loago-worker/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const (
	tlsCertPath = "../../../testdata/cert.pem"
	tlsKeyPath  = "../../../testdata/key.pem"
	secret      = "foobar"
	authScheme  = "basic"
)

var (
	testRequest = &api.RunRequest{
		Endpoints: []*api.RunRequest_Endpoint{
			{
				Url:    "http://foo.bar",
				Weight: 1,
			},
		},
		Amount:      1,
		Type:        api.RunRequest_FAKE,
		MinWaitTime: 1000,
		MaxWaitTime: 2000,
	}
)

type MockHandler struct{}

func (h *MockHandler) Run(_ *api.RunRequest, srv api.Worker_RunServer) error {
	for i := 0; i < 3; i++ {
		r := &api.EndpointResult{HttpStatusCode: 200}

		err := srv.Send(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateBufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

func ctxWithSecret(ctx context.Context, scheme string, token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

func TestServer_NoTLS_NoSecret(t *testing.T) {
	cfg := &Config{}

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(cfg, &MockHandler{}, bufconnListener)
	require.NoError(t, err)
	go func() {
		errChan <- s.Serve()
	}()
	defer s.Stop()

	select {
	case <-time.After(500 * time.Millisecond):
		break
	case err := <-errChan:
		t.Fatalf("cannot start server: %v", err)
	}

	// start client
	dialer := generateBufDialer(bufconnListener)
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	cl := api.NewWorkerClient(conn)
	stream, err := cl.Run(context.Background(), testRequest)
	require.NoError(t, err)

	var responds []*api.EndpointResult
	for {
		resp, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatalf("unknown error received while receiving gRPC message: %v", err)
			}
		}

		responds = append(responds, resp)
	}

	assert.Len(t, responds, 3)
	for _, v := range responds {
		assert.Equal(t, int32(200), v.HttpStatusCode)
	}
}

func TestServer_withTLS_NoSecret(t *testing.T) {
	cfg := &Config{
		TlsCertPath: tlsCertPath,
		TlsKeyPath:  tlsKeyPath,
	}

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(cfg, &MockHandler{}, bufconnListener)
	require.NoError(t, err)
	go func() {
		errChan <- s.Serve()
	}()
	defer s.Stop()

	select {
	case <-time.After(500 * time.Millisecond):
		break
	case err := <-errChan:
		t.Fatalf("cannot start server: %v", err)
	}

	// start client
	ctx := context.Background()
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	require.NoError(t, err)

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer conn.Close()

	cl := api.NewWorkerClient(conn)
	stream, err := cl.Run(context.Background(), testRequest)
	require.NoError(t, err)

	var responds []*api.EndpointResult
	for {
		resp, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatalf("unknown error received while receiving gRPC message: %v", err)
			}
		}

		responds = append(responds, resp)
	}

	assert.Len(t, responds, 3)
	for _, v := range responds {
		assert.Equal(t, int32(200), v.HttpStatusCode)
	}
}

func TestServer_withTLS_Unauthenticated(t *testing.T) {
	cfg := &Config{
		TlsCertPath: tlsCertPath,
		TlsKeyPath:  tlsKeyPath,
		Secret:      secret,
	}

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(cfg, &MockHandler{}, bufconnListener)
	require.NoError(t, err)
	go func() {
		errChan <- s.Serve()
	}()
	defer s.Stop()

	select {
	case <-time.After(500 * time.Millisecond):
		break
	case err := <-errChan:
		t.Fatalf("cannot start server: %v", err)
	}

	// start client
	ctx := context.Background()
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	require.NoError(t, err)

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer conn.Close()

	cl := api.NewWorkerClient(conn)
	stream, err := cl.Run(ctx, testRequest)
	require.NoError(t, err)

	resp, err := stream.Recv()
	require.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	require.True(t, ok)

	assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
}

func TestServer_withTLS_withInvalidSecret(t *testing.T) {
	cfg := &Config{
		TlsCertPath: tlsCertPath,
		TlsKeyPath:  tlsKeyPath,
		Secret:      secret,
	}

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(cfg, &MockHandler{}, bufconnListener)
	require.NoError(t, err)
	go func() {
		errChan <- s.Serve()
	}()
	defer s.Stop()

	select {
	case <-time.After(500 * time.Millisecond):
		break
	case err := <-errChan:
		t.Fatalf("cannot start server: %v", err)
	}

	// start client
	ctx := ctxWithSecret(context.Background(), authScheme, secret+"invalid")
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	require.NoError(t, err)

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer conn.Close()

	cl := api.NewWorkerClient(conn)
	stream, err := cl.Run(ctx, testRequest)
	require.NoError(t, err)

	resp, err := stream.Recv()
	require.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	require.True(t, ok)

	assert.Equal(t, codes.PermissionDenied, grpcErr.Code())
}

func TestServer_withTLS_withValidSecret(t *testing.T) {
	cfg := &Config{
		TlsCertPath: tlsCertPath,
		TlsKeyPath:  tlsKeyPath,
		Secret:      secret,
	}

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(cfg, &MockHandler{}, bufconnListener)
	require.NoError(t, err)
	go func() {
		errChan <- s.Serve()
	}()
	defer s.Stop()

	select {
	case <-time.After(500 * time.Millisecond):
		break
	case err := <-errChan:
		t.Fatalf("cannot start server: %v", err)
	}

	// start client
	ctx := ctxWithSecret(context.Background(), authScheme, secret)
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	require.NoError(t, err)

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer conn.Close()

	cl := api.NewWorkerClient(conn)
	stream, err := cl.Run(ctx, testRequest)
	require.NoError(t, err)

	var responds []*api.EndpointResult
	for {
		resp, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatalf("unknown error received while receiving gRPC message: %v", err)
			}
		}

		responds = append(responds, resp)
	}
}

func TestServer_withInvalidTLSFile(t *testing.T) {
	cfg := &Config{
		TlsCertPath: tlsCertPath + "invalid",
		TlsKeyPath:  tlsKeyPath,
		Secret:      secret,
	}

	// start server
	bufconnListener := bufconn.Listen(1024)
	s, err := newWorkerServer(cfg, &MockHandler{}, bufconnListener)

	assert.Nil(t, s)
	if assert.Error(t, err) {
		assert.IsType(t, &os.PathError{}, err)
	}
}
