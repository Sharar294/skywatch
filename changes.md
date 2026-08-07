# SkyWatch-Ops: Change Log — Analyzer Baseline + Grafana

_Date: 2026-08-07_

This round wired the first real intelligence into the pipeline and closed two of the
open items from `update.md`: **Grafana dashboards** and a **baseline anomaly detector**
in the Analyzer. The stack now runs end to end — OpenSky → Redis/Kafka → API + Analyzer →
Prometheus → Grafana — with anomaly detection instead of a placeholder log line.

---

## 1. Anomaly Detection Baseline

### New: `internal/domain/anomaly.go`
Added two domain types:
- `Anomaly` — a single flagged flight (icao24, callsign, origin, machine-readable
  `reason`, human-readable `detail`, offending `value`, and `z_score` when statistical).
- `BatchStats` — per-batch summary (total/airborne counts, anomaly count, mean and
  std-dev of velocity and altitude).

### New: `internal/service/anomaly.go`
An explainable `Detector` combining two strategies:
- **Physical rule checks** (batch-independent): impossible velocity (`>400 m/s` or `<0`),
  impossible altitude (`>16000 m` or `<-500 m`), and "on ground but moving `>120 m/s`".
- **Statistical outlier checks** (per batch, airborne flights only): flags velocity or
  altitude beyond `3σ` from the batch mean. Skipped when fewer than 8 airborne flights
  are present, so z-scores aren't computed on samples too small to trust.

Thresholds live in `NewDetector()` and are easy to tune. `Detect()` returns the anomalies
plus `BatchStats` for observability.

### New: `internal/service/anomaly_test.go`
Six unit tests: impossible velocity, impossible altitude, on-ground-but-fast, statistical
velocity outlier, no-false-positives on a clean batch, and small-batch skips statistics.

---

## 2. Analyzer Rewrite — `cmd/analyzer/main.go`

Replaced the old `// TODO: Pass to ONNX/Gorgonia` stub with a working consumer:
- Decodes each Kafka message (`[]domain.Flight`), runs the detector, and logs a concise
  batch summary plus a few example anomalies.
- Switched the Kafka reader to a **consumer group** (`skywatch-analyzer`) so offsets are
  tracked across restarts (was reading a fixed partition 0).
- Exposes a **Prometheus `/metrics`** endpoint and a `/health` endpoint on `:8081`
  (configurable via `METRICS_ADDR`).

### Metrics exposed
| Metric | Type | Meaning |
|---|---|---|
| `skywatch_analyzer_batches_total` | counter | Batches consumed from Kafka |
| `skywatch_analyzer_flights_processed_total` | counter | Individual vectors analyzed |
| `skywatch_analyzer_anomalies_total{reason}` | counter | Anomalies by reason |
| `skywatch_analyzer_anomaly_ratio` | gauge | Anomalous fraction of latest batch |
| `skywatch_analyzer_airborne_flights` | gauge | Airborne count, latest batch |
| `skywatch_analyzer_mean_velocity_mps` | gauge | Mean velocity, latest batch |
| `skywatch_analyzer_mean_altitude_m` | gauge | Mean altitude, latest batch |

---

## 3. Observability — Prometheus + Grafana

### `configs/prometheus.yml`
Added a scrape job for the analyzer (`analyzer:8081`) alongside the existing API job.

### New: Grafana provisioning under `configs/grafana/provisioning/`
- `datasources/prometheus.yml` — auto-wires the Prometheus datasource (`http://prometheus:9090`).
- `dashboards/dashboards.yml` — file-based dashboard provider.
- `dashboards/skywatch.json` — "SkyWatch-Ops Overview" dashboard: airborne flights,
  anomaly ratio gauge, flights/min, batches/min, anomalies-by-reason timeseries,
  batch mean velocity/altitude, and API request rate by endpoint.

### `docker-compose.yml`
- Added a **grafana** service (port `3000`, admin/admin, provisioning mounted,
  `grafana-data` named volume) depending on Prometheus.
- Exposed the analyzer's metrics port (`8081`) and set `METRICS_ADDR`.
- Prometheus now `depends_on` the analyzer as well as the API.

---

## Files Touched
```
NEW  internal/domain/anomaly.go
NEW  internal/service/anomaly.go
NEW  internal/service/anomaly_test.go
NEW  configs/grafana/provisioning/datasources/prometheus.yml
NEW  configs/grafana/provisioning/dashboards/dashboards.yml
NEW  configs/grafana/provisioning/dashboards/skywatch.json
MOD  cmd/analyzer/main.go        (baseline detector + Prometheus metrics)
MOD  configs/prometheus.yml      (scrape analyzer)
MOD  docker-compose.yml          (grafana service, analyzer metrics port)
```

No new external dependencies were introduced — all imports (Prometheus client,
kafka-go, etc.) were already present in `go.mod`.

---

## How to Verify (run locally)
1. **Build & test:** `go mod tidy && go build ./... && go test ./internal/service/`
2. **Run the stack:** `docker-compose up -d --build`
3. **Grafana:** open `http://localhost:3000` (admin/admin) → "SkyWatch-Ops Overview".
   Panels populate after the worker's first 60s ingestion cycle.
4. **Raw metrics:** `curl localhost:8081/metrics` (analyzer), `curl localhost:9090` (Prometheus).

## Next Steps
- Reconcile the Level-3 Kubernetes manifests (`deployments/k8s/skywatch.yaml`) with this
  working compose stack (add Kafka, Prometheus, Grafana; wire the analyzer metrics port).
- Push the detector past the baseline — labeled data + a trajectory model (ONNX/Gorgonia).
