package jobs

import "time"

type JobStatus struct {
	ID      JobID
	Command []string
	Started time.Time
	Logs    int
	Stopped *StoppedJobStatus
	Pending *PendingJobStatus
}

type StoppedJobStatus struct {
	ExitCode int
	Stopped  time.Time
}

type PendingJobStatus struct {
	CPUPercentage float32
}
