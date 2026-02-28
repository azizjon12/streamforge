package queue

import (
	"context"
	"errors"
	"time"

	"github.com/azizjon12/streamforge/internal/api"
)

var ErrQueueClosed = errors.New("queue closed")

type MemoryQueue struct {
	ch chan api.StreamJob
}

func NewMemoryQueue(buffer int) *MemoryQueue {
	return &MemoryQueue{ch: make(chan api.StreamJob, buffer)}
}

func (q *MemoryQueue) Enqueue(ctx context.Context, job api.StreamJob) error {
	select {
	case q.ch <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Dequeue blocks until a job arrives or ctx is cancelled
func (q *MemoryQueue) Dequeue(ctx context.Context) (api.StreamJob, error) {
	select {
	case job, ok := <-q.ch:
		if !ok {
			return api.StreamJob{}, ErrQueueClosed
		}
		return job, nil
	case <-ctx.Done():
		return api.StreamJob{}, ctx.Err()
	}
}

func (q *MemoryQueue) Close() {
	close(q.ch)
}

// Optional helper to avoid tight loops in worker
func SleepBackoff(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
