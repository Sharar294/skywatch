# SkyWatch-Ops: Progress & Context Update

## 🏗️ What We Started With
- **Project Scaffolding:** Standard Go project layout (`cmd/`, `internal/`, `deployments/`, etc.) generated via `setup.sh`.
- **OAuth Connectivity:** Successfully proved we could authenticate with the OpenSky Network using the `clientcredentials` flow (`test.go`).
- **Initial Planning:** Defined the 3-level progression plan in `README.md` (Local -> Docker Compose -> Kubernetes) and outlined SRE goals in `skills.md`.
- **Kubernetes Manifests:** Drafted initial deployments and services in `deployments/k8s/skywatch.yaml`.

## 🚀 What We Just Added (Level 2: Containerization & Observability)
We successfully transitioned the project into a fully functioning, containerized microservice architecture orchestrated via Docker Compose.

### 1. Go Microservices
- **Worker (`cmd/worker/main.go`):** The ingestion engine. Fetches live vectors from OpenSky via OAuth, saves the latest state to Redis, and publishes the stream to Kafka.
- **API (`cmd/api/main.go`):** The user-facing gateway. Retrieves the latest flight state from Redis and serves it as JSON. Now instrumented with **Prometheus** metrics.
- **Analyzer (`cmd/analyzer/main.go`):** The ML inference foundation. Acts as a Kafka Consumer reading the flight streams, ready for Gorgonia/ONNX integration.
- **Shared Store (`internal/service/store.go`):** The Redis client implementing producer/consumer logic for the Worker and API.

### 2. Infrastructure & Tooling
- **Kafka (Apache KRaft):** Added a single-node Apache Kafka `3.8.1` broker running in KRaft mode (no Zookeeper) to handle high-throughput telemetry streams.
- **Redis:** Implemented the "Shared Brain" using `redis:7-alpine`.
- **Prometheus:** Added a Prometheus server and `configs/prometheus.yml` to scrape the new `/metrics` endpoint on the Go API, tracking HTTP request counts.

### 3. Docker Orchestration
- **Multi-stage Dockerfile (`deployments/Dockerfile`):** Built highly optimized, secure images. Uses a shared Go `1.26-alpine` builder stage, then copies the compiled binaries into 20MB `gcr.io/distroless/static-debian12` final images.
- **Docker Compose (`docker-compose.yml`):** Networked the entire platform together so `api`, `worker`, `analyzer`, `redis-service`, `kafka`, and `prometheus` can all run seamlessly with a single `docker-compose up -d --build` command.

### 4. Overcoming SRE/DevOps Hurdles
- **Go Versioning & Network Hangs:** Successfully bypassed WSL2 MTU network hangs and Go proxy module errors by using **ephemeral Docker containers** as disposable toolchains to cleanly download dependencies (`kafka-go`, `prometheus/client_golang`) and execute `go mod tidy`.

## ⏭️ Next Steps
1. **Dashboards:** Add Grafana to `docker-compose.yml` to visualize the Prometheus metrics.
2. **Kubernetes (Level 3):** Move the working container stack back into `skywatch.yaml` and test in Minikube/Kind.
3. **Machine Learning:** Integrate Go-based ML libraries into the Analyzer to predict trajectory anomalies.