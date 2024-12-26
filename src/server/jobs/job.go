package jobs

import (
	"bufio"
	"io"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

type Job struct {
	sync.Mutex
	Id            JobID
	Command       []string
	cmd           *exec.Cmd
	cond          *sync.Cond
	isReadingDone bool
	logs          []string
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

func (job *Job) read(pipe io.ReadCloser) {
	reader := bufio.NewReader(pipe)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			job.Lock()
			job.isReadingDone = true
			job.Unlock()
			job.cond.Broadcast()
			break
		}
		job.Lock()
		job.logs = append(job.logs, line)
		job.Unlock()
		job.cond.Broadcast()
	}
}

func (job *Job) Status() JobStatus {
	job.Lock()
	defer job.Unlock()
	js := JobStatus{ID: job.Id, Logs: len(job.logs), Command: job.Command}
	if job.cmd.ProcessState.Exited() {
		js.Stopped = &StoppedJobStatus{ExitCode: job.cmd.ProcessState.ExitCode()}
	} else {
		js.Pending = &PendingJobStatus{CPUPercentage: 1.0} // TODO: implement
	}
	return js
}

// Returning 0 length indicates that there are no more logs to return
func (job *Job) GetLogs(start, maxCount int) []string {
	job.Lock()
	defer job.Unlock()
	for start <= len(job.logs) && !job.isReadingDone {
		job.cond.Wait()
	}
	return job.logs[start:min(start+maxCount, len(job.logs))]
}

func newJob(command []string) (*Job, error) {
	cmd := exec.Command(command[0], command[1:]...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	j := &Job{Id: JobID(uuid.New().String()), cmd: cmd}
	j.cond = sync.NewCond(j)
	go j.read(pipe)
	return j, nil
}
