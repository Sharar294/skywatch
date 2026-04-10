package main

import (
	"context" // Added this for Redis operations
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"skywatch/internal/service"
)

func main() {
	// 1. Signal Handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Println("🛰  SkyWatch-Worker: Starting Ingestor Engine...")

	// 2. Initialize Services
	client := service.NewOpenSkyClient()
	// Using the internal K8s DNS name we defined in the manifest
	store := service.NewStore("redis-service:6379")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// 3. Execution Context
	// We use a background context for the lifecycle of the worker
	ctx := context.Background()

	go func() {
		for {
			select {
			case <-ticker.C:
				// A. Fetch from OpenSky
				flights, err := client.FetchFlights()
				if err != nil {
					fmt.Printf("❌ [%s] Ingest Error: %v\n", time.Now().Format("15:04:05"), err)
					continue
				}

				fmt.Printf("✅ [%s] Ingested %d flight vectors\n", time.Now().Format("15:04:05"), len(flights))

				// B. Save to Redis (The "Shared Brain")
				err = store.SaveLatestFlights(ctx, flights)
				if err != nil {
					fmt.Printf("❌ [%s] Redis Save Error: %v\n", time.Now().Format("15:04:05"), err)
				} else {
					fmt.Printf("💾 [%s] Successfully synced to Redis\n", time.Now().Format("15:04:05"))
				}

			case <-stop:
				return
			}
		}
	}()

	<-stop
	fmt.Println("\n🛑 SkyWatch-Worker: Shutting down gracefully...")
}
