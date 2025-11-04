.PHONY: build run test clean docker-build docker-run dev

# Build the application
build:
	go build -o bin/weight-tracker ./cmd/server

# Run the application locally
run: build
	./bin/weight-tracker

# Development mode with hot reload
dev:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t weight-tracker:latest .

# Run with Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Database migration check
migrate-check:
	ls -la migrations/

# Full build test
build-test: clean deps test build
	@echo "Build and test completed successfully!"

# Production build
build-prod:
	CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/weight-tracker ./cmd/server

# Check if server is running
health:
	curl -f http://localhost:8080/health || exit 1

# Quick start (build and run)
start: build
	@echo "Starting Weight Tracker on http://localhost:8080"
	./bin/weight-tracker