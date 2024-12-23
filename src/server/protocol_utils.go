package main

import (
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
)

func startedJobStatusFromJob(job *jobs.Job) *teleportproto.StartedTask {
	return &teleportproto.StartedTask{Id: &teleportproto.TaskId{Uuid: string(job.Id)}}
}

func stoppedJobStatusFromStoppedJob(job *jobs.Job) *teleportproto.StoppedTask {
	return nil // TODO: Implement this method
	//return &teleportproto.StoppedTask{ErrorCode: int32(job.cmd.ProcessState.ExitCode())}
}

func jobStatusFromJob(j *jobs.Job) *teleportproto.Status {
	// result := teleportproto.Status{
	// 	Id: &teleportproto.TaskId{Uuid: string(j.Id)},
	// }
	// if j.cmd.ProcessState.Exited() {
	// 	result.TaskStatus = &teleportproto.Status_Stopped{Stopped: stoppedJobStatusFromStoppedJob(j)}
	// } else {
	// 	result.TaskStatus = &teleportproto.Status_Started{Started: startedJobStatusFromJob(j)}
	// }
	// return &result
	return nil // TODO: Implement this method
}
