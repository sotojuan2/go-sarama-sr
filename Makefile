# Go Sarama Schema Registry Producer - Makefile
.PHONY: help build build-all clean test test-integration proto lint fmt vet run-continuous run-enhanced run-robust run-basic docker-build docker-run dev-setup

# Default target
help: ## Show this help message
	@echo "Go Sarama Schema Registry Producer"
	@echo "=================================="
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build Configuration
BIN_DIR := bin
BUILD_FLAGS := -ldflags="-s -w"
GO_VERSION := 1.21

# Applications
APPS := continuous_producer enhanced_producer robust_producer producer performance_test quick_test

## Development Commands

dev-setup: ## Set up development environment (install tools, generate proto)
	@echo "🔧 Setting up development environment..."
	@go version || (echo "❌ Go $(GO_VERSION)+ required" && exit 1)
	@command -v protoc >/dev/null 2>&1 || (echo "❌ protoc required (install via package manager)" && exit 1)
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(MAKE) proto
	@echo "✅ Development environment ready!"

proto: ## Generate Protocol Buffer Go files
	@echo "🔧 Generating Protocol Buffer files..."
	@mkdir -p pb
	@protoc --go_out=./pb --go_opt=paths=source_relative pb/shoe.proto
	@echo "✅ Protocol Buffer files generated"

## Build Commands

build: ## Build all applications
	@echo "🔨 Building all applications..."
	@mkdir -p $(BIN_DIR)
	@for app in $(APPS); do \
		echo "  📦 Building $$app..."; \
		go build $(BUILD_FLAGS) -o $(BIN_DIR)/$$app ./cmd/$$app; \
	done
	@echo "✅ All applications built successfully!"

build-continuous: ## Build continuous producer (MVP)
	@echo "🔨 Building continuous producer..."
	@mkdir -p $(BIN_DIR)
	@go build $(BUILD_FLAGS) -o $(BIN_DIR)/continuous_producer ./cmd/continuous_producer
	@echo "✅ Continuous producer built: $(BIN_DIR)/continuous_producer"

build-robust: ## Build robust producer (Enterprise)
	@echo "🔨 Building robust producer..."
	@mkdir -p $(BIN_DIR)
	@go build $(BUILD_FLAGS) -o $(BIN_DIR)/robust_producer ./cmd/robust_producer
	@echo "✅ Robust producer built: $(BIN_DIR)/robust_producer"

build-enhanced: ## Build enhanced producer (Demo)
	@echo "🔨 Building enhanced producer..."
	@mkdir -p $(BIN_DIR)
	@go build $(BUILD_FLAGS) -o $(BIN_DIR)/enhanced_producer ./cmd/enhanced_producer
	@echo "✅ Enhanced producer built: $(BIN_DIR)/enhanced_producer"

build-basic: ## Build basic producer (Learning)
	@echo "🔨 Building basic producer..."
	@mkdir -p $(BIN_DIR)
	@go build $(BUILD_FLAGS) -o $(BIN_DIR)/producer ./cmd/producer
	@echo "✅ Basic producer built: $(BIN_DIR)/producer"

## Run Commands

run-continuous: build-continuous ## Run continuous producer
	@echo "🚀 Starting continuous producer..."
	@./$(BIN_DIR)/continuous_producer

run-robust: build-robust ## Run robust producer
	@echo "🚀 Starting robust producer..."
	@./$(BIN_DIR)/robust_producer

run-enhanced: build-enhanced ## Run enhanced producer
	@echo "🚀 Starting enhanced producer..."
	@./$(BIN_DIR)/enhanced_producer

run-basic: build-basic ## Run basic producer
	@echo "🚀 Starting basic producer..."
	@./$(BIN_DIR)/producer

run-quick-test: ## Run quick connectivity test
	@echo "🔍 Running quick connectivity test..."
	@go run ./cmd/quick_test

run-performance-test: ## Run performance test
	@echo "⚡ Running performance test..."
	@go run ./cmd/performance_test

## Test Commands

test: ## Run all unit tests
	@echo "🧪 Running unit tests..."
	@go test -v ./pkg/...
	@go test -v ./internal/...
	@echo "✅ Unit tests completed"

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -v ./test/...
	@go test -v ./cmd/robust_producer/ -tags=integration
	@echo "✅ Integration tests completed"

test-all: test test-integration ## Run all tests (unit + integration)

bench: ## Run benchmark tests
	@echo "⚡ Running benchmark tests..."
	@go test -bench=. ./pkg/generator
	@echo "✅ Benchmark tests completed"

## Code Quality Commands

fmt: ## Format Go code
	@echo "📝 Formatting Go code..."
	@go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet checks passed"

lint: ## Run golangci-lint (requires installation)
	@echo "🔍 Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint not installed" && exit 1)
	@golangci-lint run
	@echo "✅ Lint checks passed"

check: fmt vet ## Run basic code quality checks

## Docker Commands

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t go-sarama-producer .
	@echo "✅ Docker image built: go-sarama-producer"

docker-run: docker-build ## Run continuous producer in Docker
	@echo "🐳 Running continuous producer in Docker..."
	@docker run --env-file .env go-sarama-producer

docker-run-basic: docker-build ## Run basic producer in Docker
	@echo "🐳 Running basic producer in Docker..."
	@docker run --env-file .env go-sarama-producer ./bin/producer

docker-compose-up: ## Start with docker-compose (continuous producer)
	@echo "🐳 Starting with docker-compose..."
	@docker-compose --profile continuous up

docker-compose-down: ## Stop docker-compose services
	@echo "🐳 Stopping docker-compose services..."
	@docker-compose down

## Clean Commands

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f pb/*.pb.go
	@echo "✅ Clean completed"

clean-all: clean ## Clean everything including Docker
	@echo "🧹 Cleaning everything..."
	@docker rmi go-sarama-producer 2>/dev/null || true
	@docker-compose down --rmi all --volumes 2>/dev/null || true
	@echo "✅ Full clean completed"

## Utility Commands

deps: ## Download and tidy dependencies
	@echo "📦 Managing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

verify: ## Verify modules
	@echo "🔍 Verifying modules..."
	@go mod verify
	@echo "✅ Modules verified"

info: ## Show project information
	@echo "📊 Project Information"
	@echo "====================="
	@echo "Go version: $$(go version)"
	@echo "Project: go-sarama-sr"
	@echo "Applications: $(APPS)"
	@echo "Build directory: $(BIN_DIR)"
	@echo ""
	@echo "Structure:"
	@find cmd -name "main.go" | head -10 | while read f; do echo "  📁 $$(dirname $$f)"; done

## Development Workflow

dev: dev-setup build test ## Complete development setup (setup + build + test)
	@echo "🎉 Development environment ready!"

ci: fmt vet test build ## CI pipeline simulation (format + vet + test + build)
	@echo "✅ CI pipeline completed successfully!"

release: clean ci docker-build ## Release pipeline (clean + ci + docker)
	@echo "🎉 Release pipeline completed!"

## Environment Check

env-check: ## Check environment configuration
	@echo "🔍 Checking environment configuration..."
	@test -f .env || (echo "❌ .env file not found. Copy from .env.example" && exit 1)
	@echo "✅ .env file found"
	@echo "🔧 Environment variables:"
	@grep -E "^[A-Z_]+" .env | head -5 | while read line; do \
		var=$$(echo $$line | cut -d'=' -f1); \
		echo "  ✓ $$var"; \
	done
	@echo "  ... (and more)"

# Help text
$(shell echo "")
$(shell echo "Quick Start:")
$(shell echo "  make dev          # Complete development setup")
$(shell echo "  make run-continuous  # Run production-ready producer")
$(shell echo "  make test         # Run tests")
$(shell echo "  make docker-run   # Run in Docker")
