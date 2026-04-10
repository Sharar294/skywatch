.PHONY: build run docker-build k8s-deploy build-api build-worker build-all

build:
	go build -o bin/skywatch ./cmd/api/main.go

run:
	go run cmd/api/main.go

docker-build:
	docker build -t skywatch-api:latest -f deployments/Dockerfile .

k8s-deploy:
	kubectl apply -f deployments/k8s/api.yaml

# Build local binaries
build-api:
	go build -o bin/api ./cmd/api/main.go

build-worker:
	go build -o bin/worker ./cmd/worker/main.go

build-all: build-api build-worker

# Build Docker images
docker-api:
	docker build -t skywatch-api:latest -f deployments/Dockerfile --target api-final .

docker-worker:
	docker build -t skywatch-worker:latest -f deployments/Dockerfile --target worker-final .

docker-all: docker-api docker-worker