package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/test/bufconn"

	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const (
	secret     = "foobar"
	authScheme = "basic"
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

func handleConnClose(t *testing.T, conn *grpc.ClientConn) {
	err := conn.Close()
	if err != nil {
		t.Logf("error on closing connection: %v", err)
	}
}

func generateTLSCert() (tls.Certificate, error) {
	//priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{Certificate: [][]byte{cert}, PrivateKey: priv}, nil
}

func TestServer_NoTLS_NoSecret(t *testing.T) {
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(nil, "", &MockHandler{}, bufconnListener)
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
	defer handleConnClose(t, conn)

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
	cert, err := generateTLSCert()
	require.NoError(t, err)

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(&cert, "", &MockHandler{}, bufconnListener)
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
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer handleConnClose(t, conn)

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
	cert, err := generateTLSCert()
	require.NoError(t, err)

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(&cert, secret, &MockHandler{}, bufconnListener)
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
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer handleConnClose(t, conn)

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
	cert, err := generateTLSCert()
	require.NoError(t, err)

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(&cert, secret, &MockHandler{}, bufconnListener)
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
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer handleConnClose(t, conn)

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
	cert, err := generateTLSCert()
	require.NoError(t, err)

	// start server
	bufconnListener := bufconn.Listen(1024)
	errChan := make(chan error)
	s, err := newWorkerServer(&cert, secret, &MockHandler{}, bufconnListener)
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
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

	dialer := generateBufDialer(bufconnListener)
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer handleConnClose(t, conn)

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

	assert.Len(t, responds, 3)
	for _, v := range responds {
		assert.Equal(t, int32(200), v.HttpStatusCode)
	}

}
