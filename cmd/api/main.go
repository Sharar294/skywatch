package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. Setup Signal Handling for Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 2. Placeholder for Config Initialization
	fmt.Println("SkyWatch-Ops: Initializing Foundation...")

	// 3. Basic Router
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{\"status\": \"UP\"}")
	})

	// 4. Start Server in a Goroutine
	go func() {
		fmt.Println("📡 Ingestor listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Block until signal received
	<-stop
	fmt.Println("\n🛑 Shutting down gracefully...")
}