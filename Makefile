# Open-Telemorph-Prime Makefile

.PHONY: build run test clean docker-build docker-run help

# Variables
BINARY_NAME=open-telemorph-prime
DOCKER_IMAGE=open-telemorph-prime
VERSION=0.1.0

# Default target
all: build

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "✅ Build complete!"

# Run the application
run: build
	@echo "🚀 Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run in development mode
dev:
	@echo "🔧 Running in development mode..."
	go run main.go -config config.yaml

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Run the test script
test-integration:
	@echo "🔬 Running integration tests..."
	./test.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -rf data/
	@echo "✅ Clean complete!"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Run with Docker Compose
docker-run:
	@echo "🐳 Starting with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	@echo "🛑 Stopping Docker Compose..."
	docker-compose down

# View logs
logs:
	@echo "📋 Viewing logs..."
	docker-compose logs -f

# Create data directory
data-dir:
	@echo "📁 Creating data directory..."
	mkdir -p data

# Setup development environment
setup: deps data-dir
	@echo "⚙️ Setting up development environment..."
	@echo "✅ Setup complete!"

# Show help
help:
	@echo "Open-Telemorph-Prime Development Commands:"
	@echo ""
	@echo "  build           Build the binary"
	@echo "  run             Build and run the application"
	@echo "  dev             Run in development mode"
	@echo "  test            Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  clean           Clean build artifacts"
	@echo "  deps            Install dependencies"
	@echo "  fmt             Format code"
	@echo "  lint            Lint code"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-run      Run with Docker Compose"
	@echo "  docker-stop     Stop Docker Compose"
	@echo "  logs            View Docker logs"
	@echo "  data-dir        Create data directory"
	@echo "  setup           Setup development environment"
	@echo "  help            Show this help message"






