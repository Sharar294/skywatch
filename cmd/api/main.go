package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"skywatch/internal/service"
	"syscall"
	"time"
)

func main() {
	// 1. Setup Signal Handling for Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// 1. Use context for managing graceful shutdown.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 2. Placeholder for Config Initialization
	fmt.Println("SkyWatch-Ops: Initializing Foundation...")
	// 2. Initialize the Redis store, connecting to the same K8s service.
	store := service.NewStore("redis-service:6379")
	fmt.Println("🛰  SkyWatch-API: Connected to data store.")

	// 3. Basic Router
	// 3. Setup the HTTP router (mux)
	mux := http.NewServeMux()
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{\"status\": \"UP\"}")
	})

	// 4. Start Server in a Goroutine
	// The new endpoint to serve flight data
	http.HandleFunc("/api/v1/flights", func(w http.ResponseWriter, r *http.Request) {
		flights, err := store.GetLatestFlights(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve flight data", http.StatusInternalServerError)
			log.Printf("ERROR: Redis fetch failed: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flights)
	})

	// 4. Create and start the HTTP server.
	server := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		fmt.Println("📡 Ingestor listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		fmt.Println("📡 API server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Block until signal received
	<-stop
	fmt.Println("\n🛑 Shutting down gracefully...")
	// 5. Wait for shutdown signal, then gracefully shut down the server.
	<-ctx.Done()
	fmt.Println("\n🛑 Shutting down API server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Forced server shutdown: %v", err)
	}
}