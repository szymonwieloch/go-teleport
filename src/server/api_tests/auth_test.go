//go:build apitests
// +build apitests

package apitests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
)

func TestAuth(t *testing.T) {
	tests := []struct {
		clientSecret string
		serverSecret string
		wantOk       bool
	}{
		{clientSecret: "", serverSecret: "", wantOk: true},
		{clientSecret: "blah", serverSecret: "blah", wantOk: true},
		{clientSecret: "nope", serverSecret: "blah", wantOk: false},
		{clientSecret: "nope", serverSecret: "blah", wantOk: false},
		{clientSecret: "", serverSecret: "blah", wantOk: false},
		{clientSecret: "nope", serverSecret: "", wantOk: false},
	}

	for _, test := range tests {
		name := fmt.Sprintf("client=%s server=%s", test.clientSecret, test.serverSecret)
		t.Run(name, func(t *testing.T) {
			close := mustStartServer(t, test.serverSecret)
			defer close()

			client := mustCreateClient(t, test.clientSecret)
			defer client.close()

			req := teleportproto.Command{Command: []string{"echo", "blah"}}
			_, err := client.Start(testContext(), &req)
			gotOk := (err == nil)
			assert.Equal(t, test.wantOk, gotOk)
		})
	}
}

// Runs a short application and inspects its status and logs after it shuts down
func TestShort(t *testing.T) {
	client, close := mustCreateClientAndServer(t)
	defer close()

	req := teleportproto.Command{Command: []string{"echo", "blah"}}
	st, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st)
}

// Runs a long application and inspects its status while it is running
// Stops the application
func TestLong(t *testing.T) {
	client, close := mustCreateClientAndServer(t)
	defer close()

	req := teleportproto.Command{Command: []string{"sleep", "10"}}
	st, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st)
}

// Runs a long applicaton and inspect its logs while it is running
func TestLogs(t *testing.T) {
	client, close := mustCreateClientAndServer(t)
	defer close()

	req := teleportproto.Command{Command: []string{"sleep", "10"}}
	st, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st)
}
