# Stage 1: Shared Builder
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

# Build all binaries at once
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/worker ./cmd/worker/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/analyzer ./cmd/analyzer/main.go

# Stage 2: API Final Image
FROM gcr.io/distroless/static-debian12 AS api-final
WORKDIR /
COPY --from=builder /app/bin/api /skywatch-api
# Secrets are injected at runtime (compose env_file / K8s Secrets), never baked in.
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/skywatch-api"]

# Stage 3: Worker Final Image
FROM gcr.io/distroless/static-debian12 AS worker-final
WORKDIR /
COPY --from=builder /app/bin/worker /skywatch-worker
# Secrets are injected at runtime (compose env_file / K8s Secrets), never baked in.
USER nonroot:nonroot
# No EXPOSE needed for worker
ENTRYPOINT ["/skywatch-worker"]

# Stage 4: Analyzer Final Image
FROM gcr.io/distroless/static-debian12 AS analyzer-final
WORKDIR /
COPY --from=builder /app/bin/analyzer /skywatch-analyzer
USER nonroot:nonroot
ENTRYPOINT ["/skywatch-analyzer"]