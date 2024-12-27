// Implementation of Job - a single remote process
package jobs

import (
	"errors"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Waiting on the process to complete failed
var ErrTimeout = errors.New("Timeout")

type Job struct {
	mutex        sync.Mutex
	ID           JobID
	Command      []string
	Started      time.Time
	Stopped      time.Time
	cmd          *exec.Cmd
	logs         *logs
	killedSignal chan struct{}
}

// Stops the job and waits for it to finish.
// Thread safe.
func (job *Job) stop() error {
	err := job.kill()
	if err != nil {
		return err
	}
	// wait for the process to finish
	select {
	case <-job.killedSignal:
		return nil
	case <-time.After(5 * time.Second):
		return ErrTimeout
	}
}

// Sends a kill signal to the job.
// Does not wait for the job to finish.
// Thread safe.
func (job *Job) kill() error {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	if !job.isStopped() {
		err := job.cmd.Process.Kill()
		if err != nil {
			log.Println("Could not kill the job", job.ID, err)
			return err
		}
	}
	return nil
}

// Returns true if the process is stopped.
// NOT thread safe
func (job *Job) isStopped() bool {
	return job.Stopped != time.Time{}
}

// Returns snapshot of status of the job.
// Thread safe.
func (job *Job) Status() JobStatus {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	js := JobStatus{
		ID:      job.ID,
		Logs:    job.logs.size(),
		Command: job.Command,
		Started: job.Started,
	}

	if job.isStopped() && job.cmd.ProcessState != nil {
		js.Stopped = &StoppedJobStatus{
			ExitCode: job.cmd.ProcessState.ExitCode(),
			Stopped:  job.Stopped,
		}
	} else {
		js.Pending = &PendingJobStatus{CPUPercentage: 1.0} // TODO: implement
	}
	return js
}

// Marks the job as stopped.
// Thread safe
func (job *Job) markStopped() {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	job.Stopped = time.Now()
}

// Waits for the process to finish.
// Marks the job as stopped.
// Sends a signal to the channel when the job is stopped.
func (job *Job) wait() {
	err := job.cmd.Wait()
	log.Println("Job", job.ID, "finished")
	job.markStopped()
	if err != nil {
		log.Println("Job", job.ID, "finished with error:", err)
	}
	close(job.killedSignal) // broadcast that the job is stopped
}

// Returns logs of the job.
// start is the index of the first log entry.
// maxCount is the maximum number of log entries to return.
// Blocks until logs are available.
// Empty result indicates that there are no more logs - process stopped or closed its output channels.
func (job *Job) GetLogs(start, maxCount int) []LogEntry {
	return job.logs.get(start, maxCount)
}

// Creates a new job.
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
	id := JobID(uuid.New().String())
	j := &Job{
		ID:           id,
		cmd:          cmd,
		Started:      time.Now(),
		Command:      command,
		logs:         newLogs(stdout, stderr, id),
		killedSignal: make(chan struct{}),
	}
	go j.wait()
	return j, nil
}
