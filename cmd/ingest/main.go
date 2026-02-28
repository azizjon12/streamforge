package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/azizjon12/streamforge/internal/api"
	"github.com/azizjon12/streamforge/internal/encoding"
	"github.com/azizjon12/streamforge/internal/queue"
	"github.com/azizjon12/streamforge/internal/storage"
)

func main() {
	mux := http.NewServeMux()

	// Phase 2: local storage + in-memory store + in-memory queue
	st := storage.NewLocalStorage("./hls")
	store := api.NewJobStore()
	q := queue.NewMemoryQueue(100)

	// API handlers
	h := api.NewHandler(q, store, st)
	h.Register(mux)

	// Serve generated HLS files
	hlsFS := http.FileServer(http.Dir("./hls"))
	mux.Handle("/hls/", http.StripPrefix("/hls/", hlsFS))

	// Serve player UI
	playerFS := http.FileServer(http.Dir("./web/player"))
	mux.Handle("/player/", http.StripPrefix("/player/", playerFS))

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Start encoder worker IN-PROCESS for Phase 2 local demo.
	// In future versions, we will externalize the queue (NATS/SQS/Kafka) and run worker separately
	ff := encoding.NewFFmpeg("ffmpeg", encoding.DefaultHLSPreset())
	worker := &encoding.Worker{
		Queue:   q,
		Store:   store,
		FFmpeg:  ff,
		Logger:  log.New(os.Stdout, "worker ", log.LstdFlags),
		Timeout: 10 * time.Minute,
	}
	go func() {
		if err := worker.Run(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Ingest API listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
