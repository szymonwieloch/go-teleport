//go:build apitests
// +build apitests

package apitests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
)

func TestNoAuth(t *testing.T) {
	close := startServer(t, "")
	defer close()

	client, closeClient := createClient(t, "")
	defer closeClient()

	req := teleportproto.Command{Command: []string{"echo", "blah"}}
	st, err := client.Start(testContext(), &req)
	assert.NoError(t, err)
	t.Log(st)
}
