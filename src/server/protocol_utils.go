package main

import (
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
)

func startedJobStatus(status *jobs.RunningJobStatus) *teleportproto.StartedTask {
	return &teleportproto.StartedTask{Id: &teleportproto.TaskId{Uuid: string(status.ID)}}
}

func stoppedJobStatus(status *jobs.StoppedJobStatus) *teleportproto.StoppedTask {
	return &teleportproto.StoppedTask{ErrorCode: int32(status.ExitCode)}
}

func jobStatus(running *jobs.RunningJobStatus, stopped *jobs.StoppedJobStatus) *teleportproto.Status {
	var js jobs.JobStatus
	if running != nil {
		js = running.JobStatus
	} else {
		js = stopped.JobStatus
	}
	result := teleportproto.Status{
		Id: &teleportproto.TaskId{Uuid: string(js.ID)},
	}
	if running == nil {
		result.TaskStatus = &teleportproto.Status_Stopped{Stopped: stoppedJobStatus(stopped)}
	} else {
		result.TaskStatus = &teleportproto.Status_Started{Started: startedJobStatus(running)}
	}
	return &result

}
