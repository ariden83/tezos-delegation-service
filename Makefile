.PHONY: build run test clean docker-build docker-run

GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

APP_NAME=tezos-delegation-service

build:
	@echo "Building..."
	@go build -o $(GOBIN)/$(APP_NAME) ./cmd/tezos-delegation-service

run: build
	@echo "Running..."
	@$(GOBIN)/$(APP_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)/*

docker-build:
	@echo "Building Docker image..."
	@docker build -t tezos-delegation-service .

docker-run:
	@echo "Running with Docker..."
	@docker run -p 8080:8080 --name tezos-delegation-service tezos-delegation-service

docker-stop:
	@echo "Stopping Docker services..."
	@docker stop tezos-delegation-service || true
	@docker rm tezos-delegation-service || true

docker-compose-up:
	@echo "Building and starting all services with docker compose..."
	@docker compose build --no-cache
	@docker compose up -d

docker-compose-down:
	@echo "Stopping all services with docker compose..."
	@docker compose down

db-migrate:
	@echo "Running database migrations..."
	@./scripts/db_migrate.sh