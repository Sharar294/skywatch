package main

import (
	"context" // Added this for Redis operations
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"golang.org/x/oauth2/clientcredentials"
	"skywatch/internal/service"
)

func main() {
	// Load variables from .env file specifically pointing to configs/.env
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Println("No .env file found at configs/.env or error loading it, relying on system environment variables")
	}

	clientID := os.Getenv("OPENSKY_CLIENT_ID")
	clientSecret := os.Getenv("OPENSKY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("OPENSKY_CLIENT_ID and OPENSKY_CLIENT_SECRET must be set in your .env file or environment")
	}

	// 1. Execution Context (moved up for OAuth client)
	ctx := context.Background()

	// Configure the standard OAuth2 Client Credentials flow
	conf := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token",
	}

	oauthClient := conf.Client(ctx)
	oauthClient.Timeout = 10 * time.Second

	// 2. Signal Handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Println("🛰  SkyWatch-Worker: Starting Ingestor Engine...")

	// 3. Initialize Services
	client := service.NewOpenSkyClient(oauthClient)
	
	// Use environment variable for Redis, default to K8s DNS
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis-service:6379" 
	}
	store := service.NewStore(redisAddr)

	// Initialize Kafka Writer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka-cluster:9092"
	}
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers),
		Topic:    "flight-vectors",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Define the ingestion logic as a reusable function
	runIngestion := func() {
		// A. Fetch from OpenSky
		flights, err := client.FetchFlights()
		if err != nil {
			fmt.Printf("❌ [%s] Ingest Error: %v\n", time.Now().Format("15:04:05"), err)
			return // Use return instead of continue in a closure
		}

		fmt.Printf("✅ [%s] Ingested %d flight vectors\n", time.Now().Format("15:04:05"), len(flights))

		// B. Save to Redis (The "Shared Brain")
		err = store.SaveLatestFlights(ctx, flights)
		if err != nil {
			fmt.Printf("❌ [%s] Redis Save Error: %v\n", time.Now().Format("15:04:05"), err)
		} else {
			fmt.Printf("💾 [%s] Successfully synced to Redis\n", time.Now().Format("15:04:05"))
		}

		// C. Publish to Kafka for Analyzer
		flightsJSON, err := json.Marshal(flights)
		if err != nil {
			fmt.Printf("❌ [%s] JSON Marshal Error for Kafka: %v\n", time.Now().Format("15:04:05"), err)
		} else {
			err = kafkaWriter.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte(fmt.Sprintf("flights-%d", time.Now().Unix())),
					Value: flightsJSON,
				},
			)
			if err != nil {
				fmt.Printf("❌ [%s] Kafka Publish Error: %v\n", time.Now().Format("15:04:05"), err)
			} else {
				fmt.Printf("📤 [%s] Published to Kafka topic 'flight-vectors'\n", time.Now().Format("15:04:05"))
			}
		}
	}

	go func() {
		runIngestion() // Fire immediately on startup
		for {
			select {
			case <-ticker.C:
				runIngestion() // Run again every 60 seconds
				
			case <-stop:
				return
			}
		}
	}()

	<-stop
	fmt.Println("\n🛑 SkyWatch-Worker: Shutting down gracefully...")
}
