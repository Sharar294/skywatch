# Stage 1: Build the binary
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o skywatch ./cmd/api/main.go

# Stage 2: Final Production Image
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copy the binary from the builder
COPY --from=builder /app/skywatch /skywatch
COPY configs/.env /configs/.env

# Non-root user for security
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/skywatch"]