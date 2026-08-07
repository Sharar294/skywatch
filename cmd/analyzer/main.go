package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"

	"skywatch/internal/domain"
	"skywatch/internal/service"
)

// --- Prometheus metrics ---
var (
	batchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "skywatch_analyzer_batches_total",
		Help: "Total number of flight-vector batches consumed from Kafka",
	})
	flightsProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "skywatch_analyzer_flights_processed_total",
		Help: "Total number of individual flight vectors analyzed",
	})
	anomaliesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skywatch_analyzer_anomalies_total",
		Help: "Total anomalies detected, partitioned by reason",
	}, []string{"reason"})
	anomalyRatio = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "skywatch_analyzer_anomaly_ratio",
		Help: "Fraction of flights flagged as anomalous in the most recent batch",
	})
	airborneGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "skywatch_analyzer_airborne_flights",
		Help: "Number of airborne flights in the most recent batch",
	})
	meanVelocityGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "skywatch_analyzer_mean_velocity_mps",
		Help: "Mean ground velocity (m/s) of airborne flights in the most recent batch",
	})
	meanAltitudeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "skywatch_analyzer_mean_altitude_m",
		Help: "Mean barometric altitude (m) of airborne flights in the most recent batch",
	})
)

func main() {
	fmt.Println("🧠 SkyWatch-Analyzer: Starting ML Inference Engine...")

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka-cluster:9092"
	}

	metricsAddr := os.Getenv("METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":8081"
	}

	topic := "flight-vectors"

	// Expose Prometheus metrics + a health endpoint for K8s probes.
	go serveMetrics(metricsAddr)

	detector := service.NewDetector()

	// Setup Kafka Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBrokers},
		Topic:    topic,
		GroupID:  "skywatch-analyzer", // consumer group so offsets are tracked
		MinBytes: 10e3,                // 10KB
		MaxBytes: 10e6,                // 10MB
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
		processBatch(detector, m.Value)
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

// processBatch decodes one Kafka message into flights, runs the baseline
// anomaly detector, updates metrics, and logs a concise summary.
func processBatch(detector *service.Detector, payload []byte) {
	var flights []domain.Flight
	if err := json.Unmarshal(payload, &flights); err != nil {
		log.Printf("❌ Failed to decode batch (%d bytes): %v\n", len(payload), err)
		return
	}

	anomalies, stats := detector.Detect(flights)

	// Update Prometheus metrics.
	batchesTotal.Inc()
	flightsProcessedTotal.Add(float64(stats.TotalFlights))
	airborneGauge.Set(float64(stats.AirborneCount))
	meanVelocityGauge.Set(stats.MeanVelocity)
	meanAltitudeGauge.Set(stats.MeanAltitude)
	if stats.TotalFlights > 0 {
		anomalyRatio.Set(float64(stats.AnomalyCount) / float64(stats.TotalFlights))
	} else {
		anomalyRatio.Set(0)
	}
	for _, a := range anomalies {
		anomaliesTotal.WithLabelValues(a.Reason).Inc()
	}

	ts := time.Now().Format("15:04:05")
	fmt.Printf("✅ [%s] Batch: %d flights (%d airborne) | %d anomalies\n",
		ts, stats.TotalFlights, stats.AirborneCount, stats.AnomalyCount)

	// Log a few example anomalies so the stream is readable without Grafana.
	const maxLog = 5
	for i, a := range anomalies {
		if i >= maxLog {
			fmt.Printf("   … and %d more\n", len(anomalies)-maxLog)
			break
		}
		fmt.Printf("   🚨 %s [%s] %s — %s\n", a.ICAO24, a.Callsign, a.Reason, a.Detail)
	}
}

func serveMetrics(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"skywatch-analyzer"}`))
	})
	fmt.Printf("📊 Analyzer metrics listening on %s/metrics\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Metrics server error: %v\n", err)
	}
}
