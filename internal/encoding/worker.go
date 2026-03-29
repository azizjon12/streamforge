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

type Uploader interface {
	UploadDir(ctx context.Context, jobID string, localDir string) error
}

type Worker struct {
	Queue    Dequeuer
	Store    StatusUpdater
	FFmpeg   *FFmpeg
	Uploader Uploader
	Logger   *log.Logger
	Timeout  time.Duration
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

		if _, ok := w.Store.UpdateStatus(job.ID, api.StatusProcessing, ""); !ok {
			continue
		}

		start := time.Now()
		jobCtx, cancel := context.WithTimeout(ctx, w.Timeout)

		_, encErr := w.FFmpeg.ToHLS(jobCtx, job.Input, job.OutputDir)
		if encErr != nil {
			cancel()

			w.Store.UpdateStatus(job.ID, api.StatusFailed, encErr.Error())
			if w.Logger != nil {
				w.Logger.Printf("job=%s failed during encoding: %v", job.ID, encErr)
			}

			continue
		}

		// Upload after encoding succeeded
		if w.Uploader != nil {
			if err := w.Uploader.UploadDir(jobCtx, job.ID, job.OutputDir); err != nil {
				cancel()

				w.Store.UpdateStatus(job.ID, api.StatusFailed, "upload failed: "+err.Error())
				if w.Logger != nil {
					w.Logger.Printf("job=%s failed during upload: %v", job.ID, err)
				}

				continue
			}
		}

		cancel()

		w.Store.UpdateStatus(job.ID, api.StatusReady, "")
		if w.Logger != nil {
			w.Logger.Printf("job=%s ready (%s)", job.ID, time.Since(start).Round(time.Millisecond))
		}
	}
}
