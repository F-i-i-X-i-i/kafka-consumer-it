# Kafka Consumer - Makefile

.PHONY: build run test lint clean docker-up docker-down proto help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=consumer
MAIN_PATH=./cmd/consumer

# Docker
DOCKER_COMPOSE=docker compose

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BINARY_NAME) $(MAIN_PATH)

run: ## Run locally
	$(GOCMD) run $(MAIN_PATH)

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

docker-up: ## Start docker-compose
	$(DOCKER_COMPOSE) up -d --build

docker-down: ## Stop docker-compose
	$(DOCKER_COMPOSE) down

docker-logs: ## View docker logs
	$(DOCKER_COMPOSE) logs -f

proto: ## Generate protobuf code
	PATH=$$PATH:$$(go env GOPATH)/bin protoc --go_out=. --go_opt=paths=source_relative proto/*.proto

# Health check endpoint
health: ## Check health endpoint
	curl -s http://localhost:8080/health | jq .

# Send test message
send-test: ## Send a test message
	curl -X POST http://localhost:8080/send \
		-H "Content-Type: application/json" \
		-d '{"id":"test-1","command":"resize","image_url":"https://example.com/test.jpg","parameters":{"width":100,"height":100}}'
