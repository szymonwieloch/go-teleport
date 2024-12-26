package jobs

import (
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	sync.Mutex
	Id      JobID
	Command []string
	Started time.Time
	cmd     *exec.Cmd
	logs    *logs
}

func (job *Job) stop() error {
	job.Lock()
	defer job.Unlock()
	if !job.cmd.ProcessState.Exited() {
		err := job.cmd.Process.Kill()
		if err != nil {
			return err
		}
		err = job.cmd.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}

func (job *Job) Status() JobStatus {
	job.Lock()
	defer job.Unlock()
	js := JobStatus{
		ID:      job.Id,
		Logs:    job.logs.size(),
		Command: job.Command,
		Started: job.Started,
	}
	if job.cmd.ProcessState.Exited() {
		js.Stopped = &StoppedJobStatus{
			ExitCode: job.cmd.ProcessState.ExitCode(),
			Stopped:  time.Now(), // TODO: fix it
		}
	} else {
		js.Pending = &PendingJobStatus{CPUPercentage: 1.0} // TODO: implement
	}
	return js
}

func (job *Job) GetLogs(start, maxCount int) []LogEntry {
	return job.logs.get(start, maxCount)
}

func newJob(command []string) (*Job, error) {
	cmd := exec.Command(command[0], command[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	j := &Job{Id: JobID(uuid.New().String()), cmd: cmd, Started: time.Now(), Command: command, logs: newLogs(stdout, stderr)}
	return j, nil
}
