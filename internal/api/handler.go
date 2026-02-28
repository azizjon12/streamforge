package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/azizjon12/streamforge/internal/storage"
)

type Queue interface {
	Enqueue(ctx context.Context, job StreamJob) error
}

type Handler struct {
	Queue   Queue
	Store   *JobStore
	Storage *storage.LocalStorage
}

func NewHandler(q Queue, store *JobStore, st *storage.LocalStorage) *Handler {
	return &Handler{Queue: q, Store: store, Storage: st}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/streams", h.CreateStream)
	mux.HandleFunc("GET /v1/streams/", h.GetStream) // /v1/streams/{id}
}

func (h *Handler) CreateStream(w http.ResponseWriter, r *http.Request) {
	var req CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	req.Input = strings.TrimSpace(req.Input)
	if req.Input == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "input is required"})
		return
	}

	// Phase 2: local file input only; basic hardening
	if strings.Contains(req.Input, "..") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input path"})
		return
	}

	req.Input = filepath.Clean(req.Input)
	id := newID()

	// Ensure output dir ./hls/<id>
	if err := h.Storage.EnsurePrefix(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create output dir"})
		return
	}

	job := StreamJob{
		ID:          id,
		Input:       req.Input,
		OutputDir:   h.Storage.OutputDir(id),
		PlaylistURL: h.Storage.PlaylistPath(id),
		Status:      StatusQueued,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	h.Store.Put(job)

	if err := h.Queue.Enqueue(r.Context(), job); err != nil {
		h.Store.UpdateStatus(id, StatusFailed, err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to enqueue job"})
		return
	}

	resp := CreateStreamResponse{
		ID:          id,
		Status:      StatusQueued,
		PlaybackURL: job.PlaylistURL,
		PlayerURL:   "/player/?id=" + id,
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handler) GetStream(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/streams/")
	id = strings.TrimSpace(id)

	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"})
		return
	}

	job, ok := h.Store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, GetStreamResponse{StreamJob: job})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func newID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// fallback: timestamp-derived
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000000")))
	}
	return hex.EncodeToString(b[:])
}

// var ErrNotImplemented = errors.New("not implemented")
