# SkyWatch-Ops — Beginner Build Roadmap

A staged path to build this project yourself, one working milestone at a time.
Each milestone is a self-contained win: something you can run and see working
before moving on. Don't skip ahead — every layer assumes the one below it works.

**How to use this:** build in order. At each milestone, learn just enough of the
new concept to finish that step (don't try to master it all upfront), get it
running, commit, then move on. If you get stuck for more than ~30 minutes, that's
normal — integration debugging IS the skill you're building.

Rough total: **2–4 months part-time**, or **3–6 weeks full-time**. The times below
are "learn + build" estimates for someone new to Go and distributed systems.

---

## Milestone 0 — Setup & Tools  ·  ~1 day
Get the environment ready before writing any project code.

- [ ] Install Go (1.22+), Docker Desktop, and a code editor (VS Code + Go extension)
- [ ] Create an [OpenSky Network](https://opensky-network.org/) account and an API client (client ID + secret)
- [ ] Learn: what a terminal/shell is, basic `cd`/`ls`, and how to run `go run main.go`
- [ ] Do the official [Go Tour](https://go.dev/tour/) sections 1–3 (variables, functions, structs)

**Done when:** `go version` and `docker --version` both work, and you can run a hello-world Go file.

---

## Milestone 1 — Worker v0: Pull Flights  ·  ~3–5 days
The simplest useful thing: fetch live flight data and print it. No auth, no storage.

- [ ] Learn: Go structs, slices, `encoding/json`, making an HTTP GET with `net/http`
- [ ] Call OpenSky's `/api/states/all` endpoint and print the raw response
- [ ] Define a `Flight` struct and a mapper that turns the raw `[][]interface{}` into `[]Flight`
- [ ] Learn: type assertions and why the API returns a messy array of `any`

**Done when:** running the program prints a list of real flights (icao24, callsign, altitude, velocity).
**Maps to:** `internal/domain/flight.go`, `internal/service/mapper.go`, `internal/service/opensky.go`

---

## Milestone 2 — Add Authentication  ·  ~2–3 days
Upgrade the fetch to use your OpenSky credentials (higher rate limits, real-world auth).

- [ ] Learn: what OAuth2 "client credentials" flow is (app-to-app auth, no user login)
- [ ] Use `golang.org/x/oauth2/clientcredentials` to get a token automatically
- [ ] Load secrets from a `.env` file with `github.com/joho/godotenv` — never hardcode them
- [ ] Learn: why secrets go in `.gitignore` and never into git

**Done when:** the worker authenticates and fetches without hitting anonymous rate limits.
**Maps to:** the OAuth setup in `cmd/worker/main.go`

---

## Milestone 3 — Add Redis (Shared State)  ·  ~2–3 days
Store the latest snapshot somewhere other programs can read it.

- [ ] Learn: what Redis is (in-memory key-value store) and run it with `docker run redis`
- [ ] Use `github.com/redis/go-redis/v9` to save the flight list as JSON under a key
- [ ] Learn: why you'd set a TTL (expiry) so stale data disappears if the worker dies
- [ ] Make the worker loop on a timer (`time.Ticker`) — fetch every 60s

**Done when:** you can run `redis-cli GET latest_flights` and see your JSON.
**Maps to:** `internal/service/store.go`

---

## Milestone 4 — Build the API  ·  ~3–4 days
A web server that reads from Redis and serves flights to users.

- [ ] Learn: HTTP servers in Go (`http.ServeMux`, handlers, `ResponseWriter`)
- [ ] Add a `/v1/flights` endpoint that reads from Redis and returns JSON
- [ ] Add a `/health` endpoint (just returns "UP") — you'll need it later for Docker/K8s
- [ ] Learn: SRE basics — server timeouts and graceful shutdown on Ctrl+C (`os/signal`)

**Done when:** `curl localhost:8080/v1/flights` returns live data while the worker runs.
**Maps to:** `cmd/api/main.go`

---

## Milestone 5 — Dockerize It  ·  ~4–7 days
Package the services so they run anywhere, and wire them together.

- [ ] Learn: what a container is, and write a basic Dockerfile for one service
- [ ] Learn: multi-stage builds (compile in a big image, copy the binary into a tiny one)
- [ ] Learn: why "distroless" images are smaller and more secure
- [ ] Learn: never `COPY` secrets into an image — inject them at runtime (add a `.dockerignore`)
- [ ] Write `docker-compose.yml` to run api + worker + redis together on one network
- [ ] Learn: how compose services find each other by name (`redis-service:6379`)

**Done when:** `docker-compose up --build` starts all three and the API works.
**Maps to:** `deployments/Dockerfile`, `docker-compose.yml`, `.dockerignore`

---

## Milestone 6 — Add Kafka + Analyzer  ·  ~1–2 weeks
The hardest conceptual jump: a streaming pipeline.

- [ ] Learn: what a message broker is and why (decouple producer from consumer, buffer bursts)
- [ ] Learn: Kafka concepts — topics, partitions, offsets, consumer groups
- [ ] Run single-node Kafka in KRaft mode (no Zookeeper) via compose
- [ ] Have the worker publish each batch to a `flight-vectors` topic (`segmentio/kafka-go`)
- [ ] Build a new `analyzer` service that consumes the topic and prints what it receives

**Done when:** starting the stack, the analyzer logs batches as the worker publishes them.
**Maps to:** the Kafka writer in `cmd/worker/main.go`, `cmd/analyzer/main.go`

---

## Milestone 7 — Anomaly Detection Baseline  ·  ~3–5 days
Make the analyzer actually *do* something — the "intelligence."

- [ ] Learn: basic statistics — mean, standard deviation, z-scores (outlier detection)
- [ ] Write physical rule checks (impossible speed/altitude, "on ground but moving fast")
- [ ] Add statistical outlier checks (flag flights >3σ from the batch mean)
- [ ] Learn: writing unit tests in Go (`go test`) so you trust your logic

**Done when:** feeding it a batch with an obvious outlier flags it, and tests pass.
**Maps to:** `internal/service/anomaly.go`, `internal/service/anomaly_test.go`, `internal/domain/anomaly.go`

---

## Milestone 8 — Prometheus Metrics  ·  ~2–3 days
Make the system observable — numbers you can graph.

- [ ] Learn: what a metric is (counter vs gauge) and how Prometheus "scrapes" a `/metrics` endpoint
- [ ] Add `prometheus/client_golang` to the API and analyzer; expose counts and ratios
- [ ] Run Prometheus in compose with a config that scrapes both services

**Done when:** `localhost:9090` shows your custom `skywatch_*` metrics.
**Maps to:** the metrics code in `cmd/api/main.go` and `cmd/analyzer/main.go`, `configs/prometheus.yml`

---

## Milestone 9 — Grafana Dashboards  ·  ~2–3 days
Turn the numbers into a dashboard.

- [ ] Learn: what Grafana is and how it queries Prometheus (PromQL basics like `rate()`)
- [ ] Run Grafana in compose with a provisioned Prometheus datasource
- [ ] Build a dashboard (or provision one via JSON) with a few panels

**Done when:** `localhost:3000` shows live panels updating as the worker ingests.
**Maps to:** `configs/grafana/provisioning/`, the grafana service in `docker-compose.yml`

---

## 🎯 You've now recreated the current project.

Everything above is what exists today. Commit at each milestone so you can see your
own progress — and so a future employer can read the history.

---

## Stretch Goals (where the project is heading next)

- [ ] **Kubernetes (Level 3):** move the compose stack into K8s manifests; run it on Minikube/Kind
- [ ] **Real ML:** replace the statistical baseline with a trained trajectory/anomaly model (likely train in Python, export to ONNX)
- [ ] **Delay prediction:** the other half of the original pitch — predict arrival delays
- [ ] **CI/CD:** GitHub Actions to run `go test` and build images on every push
- [ ] **Structured logging** and multi-node Kafka/Redis for real resilience

---

### A few habits worth building from day one
- Commit small and often, with clear messages.
- When something breaks, read the *actual* error before changing code — the port
  conflicts and missing-`.env` issues in this project are solved by reading, not guessing.
- Build one service at a time and get it fully working before adding the next.
- Don't aim for "production-grade" on the first pass — aim for "runs, and I understand it."
