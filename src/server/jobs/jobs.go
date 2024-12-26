package jobs

import (
	"errors"
	"log"
	"sync"
)

var ErrNotFound = errors.New("Job was not found")

type JobID string

type Jobs struct {
	pending map[JobID]*Job
	mutex   sync.Mutex
}

func (jobs *Jobs) Create(command []string) (*Job, error) {
	j, err := newJob(command)
	if err != nil {
		log.Println("Could not create a job", err)
		return nil, err
	}
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	jobs.pending[j.Id] = j
	return j, nil
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

func NewJobs() *Jobs {
	return &Jobs{pending: make(map[JobID]*Job)}
}
