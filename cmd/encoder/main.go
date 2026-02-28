package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/azizjon12/streamforge/internal/api"
	"github.com/azizjon12/streamforge/internal/encoding"
	"github.com/azizjon12/streamforge/internal/queue"
)

func main() {
	// Phase 2 standalone worker is not very useful without a shared queue.
	// We'll still provide it for Phase 4 when the queue becomes NATS/Kafka/SQS.
	// For now, it runs with an in-memory queue just to prove behavior

	_ = api.NewJobStore()
	q := queue.NewMemoryQueue(100)

	ff := encoding.NewFFmpeg("ffmpeg", encoding.DefaultHLSPreset())
	logger := log.New(os.Stdout, "encoder", log.LstdFlags)

	worker := encoding.Worker{
		Queue:   q,
		Store:   api.NewJobStore(),
		FFmpeg:  ff,
		Logger:  logger,
		Timeout: 10 * time.Minute,
	}

	log.Println("Encoder worker started (Phase 2 local mode)")
	if err := worker.Run(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Println("cmd/encoder is reserved for Phase 4+ when we externalize the queue (NATS/SQS/Kafka) and run this worker as a separate process/pod.")

}
