# **Project: SkyWatch-Ops (Aviation Reliability & Anomaly Detection)**

## **Project Overview**
**SkyWatch-Ops** is a high-frequency, production-grade MLOps and SRE platform designed to monitor global air traffic in real-time. Built with **Golang** for maximum concurrency and low-latency execution, this system ingests live flight vectors (latitude, longitude, altitude, velocity) from the **OpenSky Network API**. It employs Machine Learning (via Go-based inference engines) to detect trajectory anomalies and predict arrival delays, focusing on the **reliability of streaming data** at scale.

---

## **The Tech Stack**

### **1. Infrastructure & Orchestration**
* **Golang (1.24+):** Leveraging goroutines and channels for high-throughput, non-blocking data ingestion.
* **Kubernetes (K8s):** Orchestrating lightweight Go binaries for self-healing and rapid scaling.
* **Docker:** Multi-stage, **distroless** builds to produce minimal (~20MB) production images.
* **Helm & Terraform:** Standardizing the deployment of the Go microservices and cloud infrastructure.

### **2. Go-Centric Backend & Data**
* **OpenSky Network API:** Live REST/State-vector stream for real-time global aviation telemetry.
* **High-Performance Routing:** Using `net/http` or `Chi` for an asynchronous, low-overhead API.
* **Inference Integration:** Leveraging **ONNX Runtime (Go bindings)** or **Gorgonia** for in-process anomaly detection.
* **Struct Tags & Validation:** Utilizing Go structs with custom tags for strict schema enforcement of telemetry data.

### **3. SRE & Observability**
* **Prometheus:** Custom Go collectors for "Inference Saturation" and "Goroutine Count."
* **Grafana:** Visualizing the "Four Golden Signals" alongside live aviation heatmaps.
* **Horizontal Pod Autoscaler (HPA):** Scaling Go pods based on CPU/Memory or custom ingestion metrics.
* **Chaos Mesh:** Testing the resiliency of Go's graceful shutdown and error-handling patterns.

---

## **Project Directory Structure**
This project follows the **Standard Go Project Layout** to ensure modularity and prevent circular dependencies.

```text
skywatch-ops/
├── cmd/                         # Entry points
│   └── api/                     # Main API application
│       └── main.go              
├── internal/                    # Private business logic
│   ├── config/                  # Env and YAML loading
│   ├── server/                  # HTTP Server & Middleware
│   ├── domain/                  # Entities (Flight, Vector) & Interfaces
│   ├── database/                # Persistence (Postgres/Redis)
│   └── service/                 # Anomaly detection & Ingestion logic
├── pkg/                         # Shared libraries (OpenSky Client)
├── api/                         # API definitions (OpenAPI/Swagger/Proto)
├── configs/                     # Configuration files (.env)
├── deployments/                 # Dockerfile & K8s manifests
├── scripts/                     # Makefile and automation scripts
├── test/                        # Integration and E2E tests
├── Makefile                     # Build/Run/Test task runner
└── go.mod                       # Dependency management