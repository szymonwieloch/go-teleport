package jobs

import (
	"errors"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("Job was not found")

type JobID string

type Job struct {
	Id    JobID
	cmd   *exec.Cmd
	mutex sync.Mutex
}

func (job *Job) stop() error {
	job.mutex.Lock()
	defer job.mutex.Unlock()
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

func (job *Job) Status() (*RunningJobStatus, *StoppedJobStatus) {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	if job.cmd.ProcessState.Exited() {
		return nil, &StoppedJobStatus{JobStatus: JobStatus{ID: job.Id}, ExitCode: job.cmd.ProcessState.ExitCode()}
	} else {
		return &RunningJobStatus{JobStatus: JobStatus{ID: job.Id}}, nil
	}
}

type Jobs struct {
	pending map[JobID]*Job
	mutex   sync.Mutex
}

func (jobs *Jobs) Create(command []string) (*Job, error) {
	cmd := exec.Command(command[0], command[1:]...)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	j := Job{Id: JobID(uuid.New().String()), cmd: cmd}
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	jobs.pending[j.Id] = &j
	return &j, nil
}

func (jobs *Jobs) Find(id JobID) *Job {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	return jobs.pending[id]
}

func (jobs *Jobs) Stop(id JobID) (*Job, error) {
	job := jobs.Find(id)
	if job == nil {
		return nil, ErrNotFound
	}
	err := job.stop()
	if err != nil {
		return nil, err
	}
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	delete(jobs.pending, id)
	return job, nil
}

func (jobs *Jobs) List() []*Job {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	result := make([]*Job, 0, len(jobs.pending))
	for _, job := range jobs.pending {
		result = append(result, job)
	}
	return result
}
