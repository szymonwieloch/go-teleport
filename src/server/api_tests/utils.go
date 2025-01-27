//go:build apitests
// +build apitests

package apitests

import (
	"context"
	"fmt"
	"io"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
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

// short version for typical use
func mustCreateClientAndServer(t *testing.T) (client, func()) {
	closeServer := mustStartServer(t, "")
	defer func() {
		if closeServer != nil {
			closeServer()
		}
	}()
	client := mustCreateClient(t, "")
	closeServer2 := closeServer
	closeServer = nil // prevent defer
	return client, func() {
		client.close()
		closeServer2()
	}
}

func isRecent(timestamp time.Time) bool {
	now := time.Now()

	return timestamp.Before(now) && timestamp.After(now.Add(-time.Second))
}

func isUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func checkJob(t *testing.T, status *teleportproto.JobStatus, cmd []string) {
	assert.True(t, isUUID(status.Id.Uuid))
	assert.True(t, isRecent(status.Started.AsTime()))
	assert.NotNil(t, status.Command)
	assert.Equal(t, status.Command.GetCommand(), cmd)
}

func checkStartedJob(t *testing.T, status *teleportproto.JobStatus, cmd []string) {
	checkJob(t, status, cmd)
	assert.Equal(t, uint32(0), status.Logs)
	assert.Nil(t, status.GetStopped())
	assert.NotNil(t, status.GetPending())
	assert.GreaterOrEqual(t, status.GetPending().CpuPerc, float32(0.0))
	assert.LessOrEqual(t, status.GetPending().CpuPerc, float32(100.0))
	assert.GreaterOrEqual(t, status.GetPending().Memory, float32(0.0))

}

func checkStoppedJob(t *testing.T, status *teleportproto.JobStatus, id string, cmd []string) {
	checkJob(t, status, cmd)
	assert.Equal(t, status.Id.Uuid, id)
	assert.NotNil(t, status.GetStopped())
	assert.Nil(t, status.GetPending())
	assert.True(t, isRecent(status.GetStopped().Stopped.AsTime()))
	assert.True(t, status.GetStopped().Stopped.AsTime().After(status.Started.AsTime()))
}

func startJob(t *testing.T, client client, cmd []string) *teleportproto.JobStatus {
	req := teleportproto.Command{Command: cmd}
	st, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st, cmd)
	return st
}

func drainLogs(t *testing.T, client client, id *teleportproto.JobId) {
	stream, err := client.Logs(testContext(), id)
	assert.NoError(t, err)
	for i := range 7 {
		resp, err := stream.Recv()
		if i == 6 {
			assert.Equal(t, err, io.EOF)
			break
		}
		assert.NoError(t, err)
		assert.Equal(t, resp.Text, fmt.Sprintf("Welcome %d times", i))
		assert.True(t, isRecent(resp.Timestamp.AsTime()))
		assert.Equal(t, resp.Src, teleportproto.LogSource_LS_STDOUT)
	}
}
