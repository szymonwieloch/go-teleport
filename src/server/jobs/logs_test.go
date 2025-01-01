package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSimpleLog(t *testing.T) {
	j, err := newJob([]string{"echo", "blah", "uf", "uf!"}, nil)
	defer j.stop()
	assert.NoError(t, err)
	logs := j.GetLogs(0, 5)
	t.Log(logs)
	assert.Equal(t, len(logs), 1)
	log := logs[0]
	assert.Equal(t, log.Line, "blah uf uf!")
	assert.True(t, log.Stdout)
	assert.WithinRange(t, log.Timestamp, time.Now().Add(-time.Second), time.Now())
}

func TestGetLogOutsideOfRange(t *testing.T) {
	j, err := newJob([]string{"echo", "blah", "uf", "uf!"}, nil)
	defer j.stop()
	assert.NoError(t, err)
	logs := j.GetLogs(1, 5)

	assert.Equal(t, len(logs), 0)
}
