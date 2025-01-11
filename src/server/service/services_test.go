package service

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/szymonwieloch/go-teleport/server/jobs"
)

func TestPendingJobStatus(t *testing.T) {
	then := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

	cmd := []string{"blah", "blue"}
	id := "36e48d8e-a44f-4f2e-803e-8353355ded6d"

	status := jobs.JobStatus{
		ID:      jobs.JobID(id),
		Command: cmd,
		Started: then,
		Logs:    23,
		Stopped: nil,
		Pending: &jobs.PendingJobStatus{
			CPUPercentage: 34.0,
			Memory:        123456789.0,
		},
	}
	grpcStatus := jobStatus(status)
	assert.NotNil(t, grpcStatus)

	assert.Equal(t, grpcStatus.Id.Uuid, id)
	assert.True(t, slices.Equal(grpcStatus.Command.Command, cmd))
	assert.Equal(t, grpcStatus.Logs, uint32(23))
	assert.Equal(t, grpcStatus.Started.AsTime(), then)
	assert.Nil(t, grpcStatus.GetStopped())
	assert.NotNil(t, grpcStatus.GetPending())
	assert.Equal(t, grpcStatus.GetPending().CpuPerc, float32(34.0))
	assert.Equal(t, grpcStatus.GetPending().Memory, float32(123456789.0))
}

func TestStoppedJobStatus(t *testing.T) {
	then := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

	later := time.Date(
		2009, 11, 17, 25, 34, 58, 651387237, time.UTC)

	cmd := []string{"blah", "blue"}
	id := "36e48d8e-a44f-4f2e-803e-8353355ded6d"

	status := jobs.JobStatus{
		ID:      jobs.JobID(id),
		Command: cmd,
		Started: then,
		Logs:    23,
		Pending: nil,
		Stopped: &jobs.StoppedJobStatus{
			ExitCode: 2,
			Stopped:  later,
		},
	}
	grpcStatus := jobStatus(status)
	assert.NotNil(t, grpcStatus)

	assert.Equal(t, grpcStatus.Id.Uuid, id)
	assert.True(t, slices.Equal(grpcStatus.Command.Command, cmd))
	assert.Equal(t, grpcStatus.Logs, uint32(23))
	assert.Equal(t, grpcStatus.Started.AsTime(), then)
	assert.NotNil(t, grpcStatus.GetStopped())
	assert.Nil(t, grpcStatus.GetPending())
	assert.Equal(t, grpcStatus.GetStopped().Stopped.AsTime(), later)
	assert.Equal(t, grpcStatus.GetStopped().ErrorCode, int32(2))
}

func TestValidSecret(t *testing.T) {
	assert.False(t, validSecret([]string{}, "password"))
	assert.False(t, validSecret([]string{"", ""}, "password"))
	assert.False(t, validSecret([]string{"Bearer Ole!"}, "password"))
	assert.True(t, validSecret([]string{"Bearer password"}, "password"))
}
