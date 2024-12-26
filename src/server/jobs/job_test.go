package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobStatus(t *testing.T) {
	js, err := newJob([]string{"echo", "hello"})
	assert.NoError(t, err)

	assert.NotNil(t, js)
	status := js.Status()
	assert.Equal(t, js.Id, status.ID)
	assert.Equal(t, js.Command, status.Command)
	assert.Equal(t, js.Started, status.Started)
	assert.Nil(t, status.Stopped)
	assert.NotNil(t, status.Pending)
}
