package api

import "time"

type StreamStatus string

const (
	StatusQueued    StreamStatus = "queued"
	StausProcessing StreamStatus = "processing"
	StatusReady     StreamStatus = "ready"
	StatusFailed    StreamStatus = "failed"
	StatusDeleted   StreamStatus = "deleted"
)

type StreamJob struct {
	ID          string       `json:"id"`
	Input       string       `json:"input"`
	OutputDir   string       `json:"_"`
	PlaylistURL string       `json:"playback_url"`
	Status      StreamStatus `json:"status"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type CreateStreamRequest struct {
	Input string `json:"input"`
}

type CreateStreamResponse struct {
	ID          string       `json:"id"`
	Status      StreamStatus `json:"status"`
	PlaybackURL string       `json:"playback_url"`
	PlayerURL   string       `json:"player_url"`
}

type GetStreamResponse struct {
	StreamJob
}
