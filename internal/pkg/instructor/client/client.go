package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/dkorittki/loago/pkg/api/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// AuthSchemeBasic is used as the authentication method description
const AuthSchemeBasic = "basic"

// Worker represents the configuration and connection of a Worker.
type Worker struct {
	Adress      string
	Port        int
	Certificate *tls.Certificate
	Secret      string
	dialer      func(context.Context, string) (net.Conn, error)
}

type ActionFunc func(context.Context, *zerolog.Logger, api.WorkerClient) error

// Client is a instructor client.
type Client struct {
	Workers  []Worker
	Actions  map[string]ActionFunc
	certPool *x509.CertPool
}

func NewClient() *Client {
	var client Client

	client.certPool = x509.NewCertPool()
	client.Actions = make(map[string]ActionFunc)

	return &client
}

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

	c.Workers = append(c.Workers, w)

	return nil
}

func (c *Client) AddAction(key string, action ActionFunc) {
	c.Actions[key] = action
}

func (c *Client) ExecuteAction(
	ctx context.Context,
	key string,
	logger *zerolog.Logger) error {

	var clients []api.WorkerClient

	for _, w := range c.Workers {
		opts := []grpc.DialOption{grpc.WithBlock()}

		if w.Certificate != nil {
			creds := credentials.NewClientTLSFromCert(c.certPool, "")
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		if w.dialer != nil {
			opts = append(opts, grpc.WithContextDialer(w.dialer))
		}

		conn, err := grpc.DialContext(ctx,
			fmt.Sprintf("%s:%d", w.Adress, w.Port), opts...)

		if err != nil {
			return &DialError{err}
		}

		defer conn.Close()

		client := api.NewWorkerClient(conn)

		if w.Secret != "" {
			ctx = ctxWithSecret(ctx, AuthSchemeBasic, w.Secret)
		}

		clients = append(clients, client)
	}

	for _, client := range clients {
		err := c.Actions[key](ctx, logger, client)
		if err != nil {
			return &ActionError{err}
		}
	}

	return nil
}

func ctxWithSecret(
	ctx context.Context,
	scheme string,
	token string) context.Context {

	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

// func newClient(workers []*Worker, dialer func(context.Context, string) (net.Conn, error)) (*Client, error) {
// 	var (
// 		client *Client
// 	)

// 	client.certPool = &x509.CertPool{}

// 	// build certpool containing all certificates for every worker
// 	for _, w := range workers {
// 		if w.Certificate != nil {
// 			for _, c := range w.Certificate.Certificate {
// 				if !client.certPool.AppendCertsFromPEM(c) {
// 					return nil, &CertificateDecodeError{}
// 				}
// 			}
// 		}
// 	}

// 	return client, nil

// 	// 		creds := credentials.NewClientTLSFromCert(cp, "")
// 	// 		opts = append(opts, grpc.WithTransportCredentials(creds))
// 	// 	} else {
// 	// 		opts = append(opts, grpc.WithInsecure())
// 	// 	}
// 	// }

// 	// for _, w := range workers {
// 	// 	worker := w
// 	// 	cp := x509.NewCertPool()
// 	// 	opts := []grpc.DialOption{grpc.WithBlock()}

// 	// 	// use TLS if certificate is set for worker
// 	// 	if w.Certificate != nil {
// 	// 		for _, c := range w.Certificate.Certificate {
// 	// 			if !cp.AppendCertsFromPEM(c) {
// 	// 				return nil, &CertificateDecodeError{}
// 	// 			}
// 	// 		}
// 	// 		creds := credentials.NewClientTLSFromCert(cp, "")
// 	// 		opts = append(opts, grpc.WithTransportCredentials(creds))
// 	// 	} else {
// 	// 		opts = append(opts, grpc.WithInsecure())
// 	// 	}

// 	// 	// use custom gRPC connection dialer if set
// 	// 	if dialer != nil {
// 	// 		opts = append(opts, grpc.WithContextDialer(dialer))
// 	// 	}

// 	// 	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

// 	// 	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", worker.Adress, worker.Port), opts...)

// 	// 	if err != nil {
// 	// 		return nil, &DialError{err}
// 	// 	}

// 	// 	worker.connection = conn
// 	// 	worker.rpcClient = api.NewWorkerClient(conn)

// 	// 	newWorkers = append(newWorkers, worker)
// 	// }

// 	// return &Client{Workers: newWorkers}, nil
// }

// // PingAllWorkers pings all workers via ping service.
// func (c *Client) PingAllWorkers(logger *zerolog.Logger) error {
// 	var returnErr error

// 	for _, w := range c.Workers {
// 		ctx := context.Background()

// 		if w.Secret != "" {
// 			ctx = ctxWithSecret(ctx, AuthSchemeBasic, w.Secret)
// 		}

// 		res, err := w.rpcClient.Ping(ctx, &api.PingRequest{})

// 		if err != nil {
// 			returnErr = errors.New("problems occured while trying to communicate with some workers")
// 			logger.Error().
// 				Str("worker", w.Alias).
// 				Err(err).
// 				Msg("cannot ping worker")
// 			continue
// 		}

// 		logger.Info().
// 			Str("worker", w.Alias).
// 			Str("response", res.Message).
// 			Msg("pinged worker")
// 	}

// 	return returnErr
// }

// // CloseConnections closes the gRPC connection to every worker of the instructor.
// func (c *Client) CloseConnections(logger zerolog.Logger) error {
// 	for _, co := range c.Configs {
// 		err := co.connection.Close()

// 		if err != nil {
// 			logger.Error().
// 				Err(err).
// 				Str("worker", co.WorkerConfig.Alias).
// 				Msg("cannot reliably close connection to worker")
// 		}
// 	}

// 	return nil
// }

// func createGrpcConnection(adress string, port int, cert string) (*grpc.ClientConn, error) {
// 	var (
// 		conn *grpc.ClientConn
// 		err  error
// 	)

// 	if cert != "" {
// 		creds, err := credentials.NewClientTLSFromFile(cert, "")

// 		if err != nil {
// 			return nil, err
// 		}

// 		conn, err = grpc.Dial(
// 			fmt.Sprintf("%s:%d", adress, port),
// 			grpc.WithTransportCredentials(creds))
// 	} else {
// 		conn, err = grpc.Dial(
// 			fmt.Sprintf("%s:%d", adress, port),
// 			grpc.WithInsecure())
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return conn, nil
// }
