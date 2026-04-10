package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"skywatch/internal/service"
)

func main() {
	fmt.Println("🚀 SkyWatch-API: Initializing gateway...")

	// 1. Connect to the Shared Brain (Redis)
	// We use the same internal K8s DNS name as the worker
	store := service.NewStore("redis-service:6379")

	// 2. Define the Router (Multiplexer)
	mux := http.NewServeMux()

	// 🟢 Health Check (For Kubernetes)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP", "service": "skywatch-api"}`))
	})

	// ✈️ Live Flight Data Endpoint (For Users)
	mux.HandleFunc("/v1/flights", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Ask Redis for the latest snapshot
		flights, err := store.GetLatestFlights(r.Context())
		if err != nil {
			// If Redis is empty (worker hasn't run yet) or down
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "Live data temporarily unavailable. Worker is syncing."}`))
			return
		}

		// Send the data to the user
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metadata": map[string]interface{}{
				"count":     len(flights),
				"timestamp": time.Now().Format(time.RFC3339),
			},
			"data": flights,
		})
	})

	// 3. Configure the HTTP Server (SRE Best Practice)
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second, // Don't let slow clients hang connections
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 4. Start Server in a Goroutine
	go func() {
		fmt.Println("📡 API listening on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// 5. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\n🛑 SkyWatch-API: Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
