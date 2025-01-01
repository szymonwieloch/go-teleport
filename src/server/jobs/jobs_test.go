package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobsFind(t *testing.T) {
	js := NewJobs(nil)
	assert.NotNil(t, js)
	j1, err := js.Create([]string{"echo", "hello"})
	defer j1.stop()
	assert.NoError(t, err)
	assert.NotNil(t, j1)

	j2 := js.Find(j1.ID)
	assert.Equal(t, j1, j2)
}

func TestCreateInvalidJob(t *testing.T) {
	js := NewJobs(nil)
	assert.NotNil(t, js)
	j, err := js.Create([]string{"barambaram"})
	assert.Nil(t, j)
	assert.Error(t, err)
	assert.Equal(t, len(js.List()), 0)
}

func TestStopJobByID(t *testing.T) {
	js := NewJobs(nil)
	j, err := js.Create([]string{"sleep", "10"})
	defer j.stop()
	assert.NoError(t, err)
	assert.NotNil(t, j)
	stopped, err := js.Stop(j.ID)
	assert.Equal(t, j, stopped)
	assert.NoError(t, err)
	assert.Equal(t, len(js.List()), 0)
}

func TestListKillAll(t *testing.T) {
	js := NewJobs(nil)
	j1, err := js.Create([]string{"sleep", "10"})
	defer j1.stop()
	assert.NoError(t, err)
	j2, err := js.Create([]string{"sleep", "10"})
	defer j2.stop()
	assert.NoError(t, err)

	// list
	pending := js.List()
	assert.Contains(t, pending, j1)
	assert.Contains(t, pending, j1)

	js.KillAll()
	time.Sleep(time.Millisecond * 100)
	assert.True(t, j1.IsStopped())
	assert.True(t, j2.IsStopped())
}
