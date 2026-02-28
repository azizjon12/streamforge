package encoding

import (
	"context"
	"log"
	"time"

	"github.com/azizjon12/streamforge/internal/api"
)

type Dequeuer interface {
	Dequeue(ctx context.Context) (api.StreamJob, error)
}

type StatusUpdater interface {
	UpdateStatus(id string, status api.StreamStatus, errMsg string) (api.StreamJob, bool)
}

type Worker struct {
	Queue   Dequeuer
	Store   StatusUpdater
	FFmpeg  *FFmpeg
	Logger  *log.Logger
	Timeout time.Duration
}

func (w *Worker) Run(ctx context.Context) error {
	if w.Timeout <= 0 {
		w.Timeout = 10 * time.Minute
	}

	for {
		job, err := w.Queue.Dequeue(ctx)
		if err != nil {
			return err
		}

		if _, ok := w.Store.UpdateStatus(job.ID, api.StausProcessing, ""); !ok {
			continue
		}

		jobCtx, cancel := context.WithTimeout(ctx, w.Timeout)
		_, encErr := w.FFmpeg.ToHLS(jobCtx, job.Input, job.OutputDir)
		cancel()

		if encErr != nil {
			w.Store.UpdateStatus(job.ID, api.StatusFailed, encErr.Error())
			if w.Logger != nil {
				w.Logger.Printf("job=%s failed: %v", job.ID, encErr)
			}

			continue
		}

		w.Store.UpdateStatus(job.ID, api.StatusReady, "")
		if w.Logger != nil {
			w.Logger.Printf("job=%s ready", job.ID)
		}
	}
}
