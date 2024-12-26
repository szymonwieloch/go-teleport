package main

import (
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
)

func jobStatus(status jobs.JobStatus) *teleportproto.JobStatus {
	result := teleportproto.JobStatus{
		Id:      &teleportproto.JobId{Uuid: string(status.ID)},
		Started: nil, // TODO
		Logs:    uint32(status.Logs),
		Command: &teleportproto.Command{Command: status.Command},
	}
	if status.Stopped != nil {
		result.Details = &teleportproto.JobStatus_Stopped{Stopped: &teleportproto.StoppedJobStatus{ErrorCode: int32(status.Stopped.ExitCode), Stopped: nil}}
	} else {
		result.Details = &teleportproto.JobStatus_Pending{Pending: &teleportproto.PendingJobStatus{CpuPerc: status.Pending.CPUPercentage}}
	}

	return &result

}
