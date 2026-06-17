# IBM Storage Virtualize Snapshot Manager - Backend Makefile

.PHONY: help build test test-verbose test-coverage clean run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the server binary
	@echo "Building server..."
	go build -o snapshot-manager ./cmd/server
	@echo "Build complete: snapshot-manager"

test: ## Run all tests
	@echo "Running tests..."
	go test ./cmd/... ./internal/... ./pkg/...

test-verbose: ## Run all tests with verbose output
	@echo "Running tests (verbose)..."
	go test -v ./cmd/... ./internal/... ./pkg/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./cmd/... ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-api: ## Run API tests only
	@echo "Running API tests..."
	go test -v ./internal/api/...

clean: ## Clean build artifacts and test files
	@echo "Cleaning..."
	rm -f snapshot-manager
	rm -f coverage.out coverage.html
	@echo "Clean complete"

run: build ## Build and run the server
	@echo "Starting server..."
	./snapshot-manager

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: fmt vet ## Run formatters and linters
	@echo "Linting complete"

all: clean lint test build ## Run clean, lint, test, and build
	@echo "All tasks complete"

# 
