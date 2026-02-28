package api

import (
	"sync"
	"time"
)

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]StreamJob
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]StreamJob)}
}

func (s *JobStore) Put(job StreamJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

func (s *JobStore) Get(id string) (StreamJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	return j, ok
}

func (s *JobStore) UpdateStatus(id string, status StreamStatus, errMsg string) (StreamJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return StreamJob{}, false
	}
	j.Status = status
	j.UpdatedAt = time.Now()
	j.Error = errMsg
	s.jobs[id] = j
	return j, true
}
