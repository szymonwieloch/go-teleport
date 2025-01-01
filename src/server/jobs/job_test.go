package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobCreateStop(t *testing.T) {
	js, err := newJob([]string{"sleep", "10000"}, nil)
	assert.NoError(t, err)
	stopTime := time.Now()
	err = js.stop()
	assert.NoError(t, err)
	stoppedTime := time.Now()
	assert.Less(t, stoppedTime, stopTime.Add(time.Second))

}

func TestJobStatus(t *testing.T) {
	js, err := newJob([]string{"sleep", "10"}, nil)
	defer js.kill()
	assert.NoError(t, err)

	assert.NotNil(t, js)
	status := js.Status()
	assert.Equal(t, js.ID, status.ID)
	assert.Equal(t, js.Command, status.Command)
	assert.Equal(t, js.Started, status.Started)
	assert.Nil(t, status.Stopped)
	assert.NotNil(t, status.Pending)

	err = js.stop()
	assert.NoError(t, err)
	status = js.Status()
	assert.Equal(t, js.ID, status.ID)
	assert.Equal(t, js.Command, status.Command)
	assert.Equal(t, js.Started, status.Started)
	assert.NotNil(t, status.Stopped)
	assert.Nil(t, status.Pending)
}
