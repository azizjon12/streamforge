package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/azizjon12/streamforge/internal/api"
	"github.com/azizjon12/streamforge/internal/encoding"
	"github.com/azizjon12/streamforge/internal/queue"
	"github.com/azizjon12/streamforge/internal/storage"
)

func main() {
	mux := http.NewServeMux()

	// Phase 2 foundation: local storage + in-memory store + in-memory queue
	st := storage.NewLocalStorage("./hls")
	store := api.NewJobStore()
	q := queue.NewMemoryQueue(100)

	// Phase 3 config
	awsRegion := os.Getenv("AWS_REGION")
	bucket := os.Getenv("STREAMFORGE_S3_BUCKET")
	prefix := os.Getenv("STREAMFORGE_S3_PREFIX")
	cloudFrontDomain := os.Getenv("STREAMFORGE_CLOUDFRONT_DOMAIN")

	// API handlers
	h := api.NewHandler(q, store, st, cloudFrontDomain, prefix)
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

	// S3 uploader
	var uploader *storage.S3Uploader
	if awsRegion != "" && bucket != "" && cloudFrontDomain != "" {
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(awsRegion))
		if err != nil {
			log.Fatalf("failed to log AWS config: %v", err)
		}

		s3Client := s3.NewFromConfig(cfg)
		uploader = storage.NewS3Uploader(bucket, prefix, s3Client)

		log.Printf("phase 3 enabled: S3 bucket=%s prefix=%s CloudFront=%s", bucket, prefix, cloudFrontDomain)
	} else {
		log.Printf("phase 3 disabled: set STREAMFORGE_S3_BUCKET and STREAMFORGE_CLOUDFRONT_DOMAIN to enable S3 + CloudFront publishing")
	}

	// Start encoder worker IN-PROCESS for Phase 2 local demo.
	// In future versions, we will externalize the queue (NATS/SQS/Kafka) and run worker separately
	ff := encoding.NewFFmpeg("ffmpeg", encoding.DefaultHLSPreset())
	workerLogger := log.New(os.Stdout, "worker", log.LstdFlags)

	worker := &encoding.Worker{
		Queue:    q,
		Store:    store,
		FFmpeg:   ff,
		Uploader: uploader,
		Logger:   workerLogger,
		Timeout:  10 * time.Minute,
	}
	go func() {
		if err := worker.Run(context.Background()); err != nil {
			log.Fatalf("worker stopped: %v", err)
		}
	}()

	log.Println("Ingest API listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
