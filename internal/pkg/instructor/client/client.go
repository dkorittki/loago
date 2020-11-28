package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/dkorittki/loago/pkg/instructor/config"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// AuthSchemeBasic is used as the authentication method description.
const AuthSchemeBasic = "basic"

type Result struct {
	URL               url.URL
	HttpStatusCode    int
	HttpStatusMessage string
	Ttfb              time.Duration
	Cached            bool
}

// Worker represents the configuration and connection of a Worker.
type Worker struct {
	Adress      string
	Port        int
	Certificate *tls.Certificate
	Secret      string
	dialer      func(context.Context, string) (net.Conn, error)
	connection  *grpc.ClientConn
}

// Client is a instructor client.
type Client struct {
	Workers  []*Worker
	certPool *x509.CertPool
}

// NewClient returns a new client.
func NewClient() *Client {
	var client Client

	client.certPool = x509.NewCertPool()

	return &client
}

// AddWorker adds a new worker to the client.
// A worker is defined by it's connection attributes,
// which are it's ip address, port, a token secret,
// a TLS certificate, and an optional dialer function, which mainly serves
// for testing purposes.
func (c *Client) AddWorker(
	adress string,
	port int,
	secret string,
	cert *tls.Certificate,
	dialer func(context.Context, string) (net.Conn, error)) error {

	var w Worker

	w.Adress = adress
	w.Port = port
	w.Secret = secret
	w.Certificate = cert
	w.dialer = dialer

	if cert != nil {
		for _, byteCert := range w.Certificate.Certificate {
			if !c.certPool.AppendCertsFromPEM(byteCert) {
				return &CertificateDecodeError{}
			}
		}
	}

	c.Workers = append(c.Workers, &w)

	return nil
}

// Connect opens a connection to every worker added to this client.
// It closes existing connections first, then creates a new one.
func (c *Client) Connect(ctx context.Context, logger *zerolog.Logger) error {
	err := c.Disconnect()

	if err != nil {
		return err
	}

	for _, w := range c.Workers {
		w.connection, err = connect(ctx, w, c.certPool)

		if err != nil {
			_ = c.Disconnect()
			return &ConnectError{err}
		}
	}

	return nil
}

// Disconnect closes connections to all workers belonging to this client.
func (c *Client) Disconnect() error {
	for _, w := range c.Workers {
		if w.connection != nil {
			err := w.connection.Close()
			w.connection = nil

			if err != nil {
				return &DisconnectError{err}
			}
		}
	}

	return nil
}

// Ping serially pings every Worker.
func (c *Client) Ping(ctx context.Context, logger *zerolog.Logger) error {
	for _, w := range c.Workers {
		ctx = ctxWithSecret(ctx, AuthSchemeBasic, w.Secret)
		client := api.NewWorkerClient(w.connection)

		res, err := client.Ping(ctx, &api.PingRequest{})

		if err != nil {
			return err
		}

		logger.Info().
			Str("response", res.Message).
			Str("worker", fmt.Sprintf("%s:%d", w.Adress, w.Port)).
			Msg("ping succeeded")
	}

	return nil
}

// Todo: Dummy function, needs implementation.
func (c *Client) Run(
	ctx context.Context,
	logger *zerolog.Logger,
	endpoints []config.InstructorEndpoint,
	results chan Result) error {

	return nil
}

// connect establishes a gRPC connection to a worker and returns the
// connection and a cancelation func for ending the connection.
func connect(ctx context.Context, w *Worker, certPool *x509.CertPool) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithBlock()}

	if w.Certificate != nil {
		creds := credentials.NewClientTLSFromCert(certPool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if w.dialer != nil {
		opts = append(opts, grpc.WithContextDialer(w.dialer))
	}

	return grpc.DialContext(ctx,
		fmt.Sprintf("%s:%d", w.Adress, w.Port), opts...)
}

// ctxWithSecret sets authorization metadata on
// the returned context using a token.
func ctxWithSecret(
	ctx context.Context,
	scheme string,
	token string) context.Context {

	if token == "" {
		return ctx
	}

	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}
