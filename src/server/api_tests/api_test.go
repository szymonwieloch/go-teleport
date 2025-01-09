//go:build apitests
// +build apitests

package apitests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
)

var shortCmd []string = []string{"echo", "blah"}
var longCmd []string = []string{"sleep", "10"}
var loggingCmd []string = []string{"bash", "-c", "for i in {0..5}; do echo Welcome $i times; sleep 0.1; done"}

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

			req := teleportproto.Command{Command: shortCmd}
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

	req := teleportproto.Command{Command: shortCmd}
	st1, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st1, shortCmd)
	time.Sleep(time.Millisecond * 100)
	// it should be done by now
	st2, err := client.GetStatus(testContext(), st1.Id)
	assert.NoError(t, err)
	checkStoppedJob(t, st2, st1.Id.Uuid, shortCmd)
	assert.Equal(t, st2.GetStopped().ErrorCode, int32(0))

	// Stop is supposed to remove the job from the internal list and free resources

	st3, err := client.Stop(testContext(), st1.Id)
	assert.NoError(t, err)
	checkStoppedJob(t, st3, st1.Id.Uuid, shortCmd)
	assert.Equal(t, st2.GetStopped().ErrorCode, int32(0))

	_, err = client.GetStatus(testContext(), st1.Id)
	assert.Error(t, err)
}

// Runs a long application and inspects its status while it is running
// Stops the application
func TestLong(t *testing.T) {
	client, close := mustCreateClientAndServer(t)
	defer close()
	req := teleportproto.Command{Command: longCmd}
	st1, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st1, longCmd)

	st2, err := client.GetStatus(testContext(), st1.Id)
	assert.NoError(t, err)
	checkStartedJob(t, st2, longCmd)

	st3, err := client.Stop(testContext(), st1.Id)
	assert.NoError(t, err)
	checkStoppedJob(t, st3, st1.Id.Uuid, longCmd)
	assert.Equal(t, st3.GetStopped().ErrorCode, int32(-1))

	_, err = client.GetStatus(testContext(), st1.Id)
	assert.Error(t, err)
}

// Runs a long applicaton and inspect its logs while it is running
func TestLogs(t *testing.T) {
	client, close := mustCreateClientAndServer(t)
	defer close()

	req := teleportproto.Command{Command: loggingCmd}
	st, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st, loggingCmd)

}

// Runs two jobs in parallel to check if there are any races
func TestParallel(t *testing.T) {
	client, close := mustCreateClientAndServer(t)
	defer close()

	req := teleportproto.Command{Command: loggingCmd}
	st1, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st1, loggingCmd)

	st2, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	checkStartedJob(t, st2, loggingCmd)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		drainLogs(t, client, st1.Id)
		wg.Done()
	}()

	go func() {
		drainLogs(t, client, st1.Id)
		wg.Done()
	}()

	wg.Wait()
}
