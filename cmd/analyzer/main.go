package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

func main() {
	fmt.Println("🧠 SkyWatch-Analyzer: Starting ML Inference Engine...")

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka-cluster:9092"
	}

	topic := "flight-vectors"

	// Setup Kafka Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBrokers},
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		fmt.Println("\n🛑 SkyWatch-Analyzer: Shutting down gracefully...")
		cancel()
	}()

	fmt.Printf("🎧 Listening for flight vectors on topic: %s\n", topic)

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break // Context canceled, exiting cleanly
			}
			log.Printf("❌ Error reading message: %v\n", err)
			continue
		}

		// TODO: Pass m.Value to ONNX / Gorgonia for anomaly detection
		fmt.Printf("✅ Received flight vector data at offset %d (size: %d bytes)\n", m.Offset, len(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}