//go:build apitests
// +build apitests

package apitests

import (
	"context"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	rootDir := path.Dir(path.Dir(path.Dir(filename)))
	allParts := []string{rootDir}
	allParts = append(allParts, elem...)
	return path.Join(allParts...)
}

func startServer(t *testing.T, secret string) func() {
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
		t.Fatalf("could not start server: %s", err)
	}
	close := srv.ServeInBackground()
	return func() {
		err := close()
		assert.NoError(t, err)
	}
}

func testContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	return ctx
}

func createClient(t *testing.T, secret string) (teleportproto.RemoteExecutorClient, func()) {

	var opts []grpc.DialOption
	if secret == "" {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	} else {
		perRPC := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: secret,
		})}
		creds, err := credentials.NewClientTLSFromFile(relativePath("certs", "ca_cert.pem"), "x.test.example.com") // TODO
		if err != nil {
			t.Fatalf("failed to load credentials: %v", err)
		}
		opts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(perRPC),
			grpc.WithTransportCredentials(creds),
		}
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		t.Fatalf("did not connect to the server: %v", err)
	}
	return teleportproto.NewRemoteExecutorClient(conn), func() { conn.Close() }
}
