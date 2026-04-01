# SkyWatch-Ops ✈️
### Real-time Aviation Reliability & Anomaly Detection Platform

**SkyWatch-Ops** is a high-frequency, production-grade MLOps and SRE platform designed to monitor global air traffic. Built with **Golang** for maximum concurrency, the system ingests live flight vectors from the **OpenSky Network API** to detect trajectory anomalies and predict arrival delays.

The project is a demonstration of **Reliable Streaming AI**, focusing on how to scale inference workloads dynamically and maintain system observability under fluctuating global flight volumes.

---

## Tech Stack
* **Language:** Go 1.26+ (Goroutines, Channels, Context)
* **Infrastructure:** Kubernetes (K8s), Docker (Multi-stage Distroless)
* **Orchestration:** Docker Compose, Helm (Planned)
* **Observability:** Prometheus, Grafana, Chaos Mesh
* **Data Source:** OpenSky Network REST/State-vector API

---

## Project Structure
Following the **Standard Go Project Layout** for modularity and scalability:

```text
skywatch-ops/
├── cmd/api/             # Entry point (main.go)
├── internal/            # Private business logic
│   ├── config/          # Environment/Config management
│   ├── domain/          # Entities (Flight/Vector) & Interfaces
│   ├── server/          # HTTP Server & Middleware
│   └── service/         # API Clients & Anomaly Logic
├── deployments/         # Dockerfile & K8s manifests
├── configs/             # .env and configuration files
├── Makefile             # Task automation
└── go.mod               # Dependency management

---

## Testing
Level 1 Local Development:
go mod init skywatch-ops
go mod tidy
make run

Level 2 Containerized (Docker):
make docker-build
docker-compose up

Level 3: Orchestrated (Kubernetes):
make k8s-deploy
kubectl get pods
kubectl port-forward service/skywatch-service 8080:8080
kubectl logs -l app=skywatch