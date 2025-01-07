//go:build apitests
// +build apitests

package apitests

import (
	"context"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"github.com/szymonwieloch/go-teleport/server/service"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
)

const address string = "localhost:1234"

// Creates a path relative to the root of this project
func relativePath(elem ...string) string {

	_, filename, _, _ := runtime.Caller(0)
	rootDir := path.Dir(path.Dir(path.Dir(path.Dir(filename))))
	allParts := []string{rootDir}
	allParts = append(allParts, elem...)
	return path.Join(allParts...)
}

func startServer(secret string) (func() error, error) {
	opts := service.ServiceOptions{
		Address: address,
	}
	if secret != "" {
		opts.AuthCert = relativePath("certs", "server_cert.pem")
		opts.AuthKey = relativePath("certs", "server_key.pem")
		opts.Secret = secret
	}
	srv, err := service.NewService(opts)
	if err != nil {
		return nil, err
	}
	close := srv.ServeInBackground()
	return close, nil
}

func mustStartServer(t *testing.T, secret string) func() {
	close, err := startServer(secret)
	if err != nil {
		t.Fatalf("could not start server: %s", err)
	}
	return func() {
		err := close()
		if err != nil {
			t.Fatalf("error while stopping server: %v", err)
		}
	}
}

func testContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	return ctx
}

type client struct {
	teleportproto.RemoteExecutorClient
	conn *grpc.ClientConn
}

func (client client) close() {
	client.conn.Close()
}

func createClient(secret string) (client, error) {
	var opts []grpc.DialOption
	if secret == "" {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	} else {
		perRPC := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: secret,
		})}
		creds, err := credentials.NewClientTLSFromFile(relativePath("certs", "ca_cert.pem"), "x.test.example.com") // TODO
		if err != nil {
			return client{}, err
		}
		opts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(perRPC),
			grpc.WithTransportCredentials(creds),
		}
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return client{}, err
	}
	return client{
		RemoteExecutorClient: teleportproto.NewRemoteExecutorClient(conn),
		conn:                 conn,
	}, nil
}

func mustCreateClient(t *testing.T, secret string) client {
	cli, err := createClient(secret)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return cli
}
