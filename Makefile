.PHONY: build run test clean install deps

# Binary name
BINARY_NAME=dns-defense-server

# Build the application
build:
	@echo "Building..."
	go build -o $(BINARY_NAME) ./cmd/server

# Build with optimizations
build-prod:
	@echo "Building for production..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/server

# Run the application (non-privileged port)
run:
	@echo "Running on port 8053..."
	go run ./cmd/server -port 8053

# Run with sudo (privileged port 53)
run-sudo:
	@echo "Running on port 53 (requires sudo)..."
	sudo go run ./cmd/server -port 53

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf logs/*.log

# Install the binary to system
install:  build-prod
	@echo "Installing to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		cp $(BINARY_NAME) /usr/local/bin/; \
		chmod 755 /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo cp $(BINARY_NAME) /usr/local/bin/; \
		sudo chmod 755 /usr/local/bin/$(BINARY_NAME); \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Create necessary directories
setup:
	@echo "Setting up directories..."
	mkdir -p logs
	mkdir -p configs

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-prod    - Build with optimizations"
	@echo "  run           - Run on port 5353"
	@echo "  run-sudo      - Run on port 53 (requires sudo)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Install dependencies"
	@echo "  clean         - Remove build artifacts"
	@echo "  install       - Install binary to system"
	@echo "  fmt           - Format code"
	@echo "  setup         - Create necessary directories"
