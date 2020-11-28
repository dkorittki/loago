package client

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/dkorittki/loago/internal/pkg/testing/fakeserver"
	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufConnBufferSize = 1024 * 1024

func newTestServer() *grpc.Server {
	s := grpc.NewServer()
	api.RegisterWorkerServer(s, &fakeserver.FakeWorkerServer{})

	return s
}

func newBufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

func assertOnWrappedError(t *testing.T, err error) {
	if err != nil {
		t.Log("error:", err.Error())
	}

	e, ok := err.(WrappableError)

	if !ok {
		assert.NoError(t, err)
		return
	}

	assertOnWrappedError(t, e.Unwrap())
}

func TestPing(t *testing.T) {
	lis1 := bufconn.Listen(bufConnBufferSize)
	lis2 := bufconn.Listen(bufConnBufferSize)
	lis3 := bufconn.Listen(bufConnBufferSize)

	dailer1 := newBufDialer(lis1)
	dailer2 := newBufDialer(lis2)
	dailer3 := newBufDialer(lis3)

	client := NewClient()
	client.AddWorker("127.0.0.1", 1234, "test123", nil, dailer1)
	client.AddWorker("127.0.0.1", 1235, "test123", nil, dailer2)
	client.AddWorker("127.0.0.1", 1236, "test123", nil, dailer3)

	server1 := newTestServer()
	server2 := newTestServer()
	server3 := newTestServer()

	go func() {
		if err := server1.Serve(lis1); err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()
	go func() {
		if err := server2.Serve(lis2); err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()
	go func() {
		if err := server3.Serve(lis3); err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()

	logger := zerolog.New(os.Stdout)

	err := client.Connect(context.Background(), &logger)
	assert.NoError(t, err)

	err = client.Ping(context.Background(), &logger)
	assertOnWrappedError(t, err)

	err = client.Disconnect()
	assert.NoError(t, err)
}
