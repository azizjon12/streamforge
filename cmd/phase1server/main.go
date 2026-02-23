package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Serve HLS segments
	hlsServer := http.FileServer(http.Dir("./hls"))
	mux.Handle("/hls/", http.StripPrefix("/hls/", hlsServer))

	// Serve player UI
	playerServer := http.FileServer(http.Dir("./web/player"))
	mux.Handle("/player/", http.StripPrefix("/player/", playerServer))

	log.Println("Phase 1 server running at http:localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
