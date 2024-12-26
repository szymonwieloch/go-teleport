package jobs

import (
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	mutex   sync.Mutex
	Id      JobID
	Command []string
	Started time.Time
	Stopped time.Time
	cmd     *exec.Cmd
	logs    *logs
}

func (job *Job) stop() error {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	if !job.isStopped() {
		err := job.cmd.Process.Kill()
		if err != nil {
			log.Println("Could not kill the job", job.Id, err)
			return err
		}
	}
	return nil
}

func (job *Job) isStopped() bool {
	return job.Stopped != time.Time{}
}

func (job *Job) Status() JobStatus {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	js := JobStatus{
		ID:      job.Id,
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

func (job *Job) wait() {
	err := job.cmd.Wait()
	log.Println("Job", job.Id, "finished")
	job.mutex.Lock()
	defer job.mutex.Unlock()
	job.Stopped = time.Now()
	if err != nil {
		log.Println("Job", job.Id, "finished with error:", err)
	}
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
	id := JobID(uuid.New().String())
	j := &Job{Id: id, cmd: cmd, Started: time.Now(), Command: command, logs: newLogs(stdout, stderr, id)}
	go j.wait()
	return j, nil
}
