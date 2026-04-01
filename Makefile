.PHONY: build run docker-build k8s-deploy

build:
	go build -o bin/skywatch ./cmd/api/main.go

run:
	go run cmd/api/main.go

docker-build:
	docker build -t skywatch-api:latest -f deployments/Dockerfile .

k8s-deploy:
	kubectl apply -f deployments/k8s/api.yaml