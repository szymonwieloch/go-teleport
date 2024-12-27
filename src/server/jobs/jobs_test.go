package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobsFind(t *testing.T) {
	js := NewJobs()
	assert.NotNil(t, js)
	j1, err := js.Create([]string{"echo", "hello"})
	assert.NoError(t, err)
	assert.NotNil(t, j1)

	j2 := js.Find(j1.ID)
	assert.Equal(t, j1, j2)
}
