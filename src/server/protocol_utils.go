package main

import (
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Maps job status as reported by a job to the gRPC equivalent
func jobStatus(status jobs.JobStatus) *teleportproto.JobStatus {
	result := teleportproto.JobStatus{
		Id:      &teleportproto.JobId{Uuid: string(status.ID)},
		Started: timestamppb.New(status.Started),
		Logs:    uint32(status.Logs),
		Command: &teleportproto.Command{Command: status.Command},
	}
	if status.Stopped != nil {
		result.Details = &teleportproto.JobStatus_Stopped{
			Stopped: &teleportproto.StoppedJobStatus{
				ErrorCode: int32(status.Stopped.ExitCode),
				Stopped:   timestamppb.New(status.Stopped.Stopped),
			},
		}
	} else {
		result.Details = &teleportproto.JobStatus_Pending{
			Pending: &teleportproto.PendingJobStatus{
				CpuPerc: status.Pending.CPUPercentage,
			},
		}
	}

	return &result

}
