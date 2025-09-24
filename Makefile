.PHONY: help build run test clean setup-certs deps docker

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  setup-certs  - Generate development certificates"
	@echo "  deps         - Download dependencies"
	@echo "  docker       - Build Docker image"
	@echo "  test-auth    - Test device authentication"

# Build the application
build:
	go build -o bin/device-assignment-api ./cmd/server

# Run the application
run: build
	./bin/device-assignment-api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Generate development certificates
setup-certs:
	./scripts/setup-certs.sh

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker:
	docker build -t device-assignment-api .

# Test device authentication
test-auth:
	go run scripts/test-client.go auth

# Test with environment variables loaded
run-with-env:
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v ^# | xargs) && ./bin/device-assignment-api; \
	else \
		echo "No .env file found. Copy env.example to .env and configure it."; \
	fi
