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

// Create creates a new job with the given command.
// Adds it to the internal collection.
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

// Find returns a job by its ID.
func (jobs *Jobs) Find(id JobID) *Job {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	return jobs.pending[id]
}

// Stop stops a job by its ID.
// Job is removed from the collection.
// On success Job instance is returned and can be used to obtain job information.
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

// Creates a snapshot of the current collection of the jobs.
func (jobs *Jobs) List() []*Job {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	result := make([]*Job, 0, len(jobs.pending))
	for _, job := range jobs.pending {
		result = append(result, job)
	}
	return result
}

// Kills all running processes.
// Does not remove processes from the collection.
// Does not wait until processes fully complete.
func (jobs *Jobs) KillAll() {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()
	for _, job := range jobs.pending {
		job.kill()
	}
}

// NewJobs creates a new collection of jobs.
func NewJobs() *Jobs {
	return &Jobs{pending: make(map[JobID]*Job)}
}
