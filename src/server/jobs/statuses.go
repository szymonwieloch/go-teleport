package jobs

type JobStatus struct {
	ID JobID
}

type StoppedJobStatus struct {
	JobStatus
	ExitCode int
}

type RunningJobStatus struct {
	JobStatus
}
