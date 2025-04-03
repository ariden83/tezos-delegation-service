.PHONY: build run test clean docker-build docker-run

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Application name
APP_NAME=tezos-delegation-service

# Build the application
build:
	@echo "Building..."
	@go build -o $(GOBIN)/$(APP_NAME) ./cmd/tezos-delegation-service

# Run the application
run: build
	@echo "Running..."
	@$(GOBIN)/$(APP_NAME)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)/*

# Build docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t tezos-delegation-service .

# Build docker image with multistage (linting and testing)
docker-build-full:
	@echo "Building Docker image with full validation (lint, test)..."
	@docker build -t tezos-delegation-service -f Dockerfile.multistage .

# Run with docker
docker-run:
	@echo "Running with Docker..."
	@docker run -p 8080:8080 --name tezos-delegation-service tezos-delegation-service

# Stop docker services
docker-stop:
	@echo "Stopping Docker services..."
	@docker stop tezos-delegation-service || true
	@docker rm tezos-delegation-service || true

# Run all services with docker compose
docker-compose-up:
	@echo "Starting all services with docker compose..."
	@docker compose up -d

# Stop all services with docker compose
docker-compose-down:
	@echo "Stopping all services with docker compose..."
	@docker compose down

# Run database migrations with sqitch
db-migrate:
	@echo "Running database migrations..."
	@./scripts/db_migrate.sh