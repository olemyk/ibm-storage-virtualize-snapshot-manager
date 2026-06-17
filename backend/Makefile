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

# Dependency management
deps-check: ## Check for outdated dependencies
	@echo "Checking Go dependencies..."
	@go list -u -m all | grep '\[' || echo "All dependencies are up to date!"

deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated. Please run tests to verify."

deps-audit: ## Run security audit on dependencies
	@echo "Running security audit..."
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	@govulncheck ./...

deps-graph: ## Show dependency graph
	@echo "Generating dependency graph..."
	@go mod graph

deps-tidy: ## Clean up dependencies
	@echo "Tidying dependencies..."
	@go mod tidy
	@go mod verify

# Container builds
docker-build: ## Build Docker image for current platform
	@echo "Building Docker image..."
	docker build -t snapshot-manager-backend:latest .

docker-build-multiarch: ## Build multi-architecture Docker image
	@echo "Building multi-architecture Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t snapshot-manager-backend:latest .

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env snapshot-manager-backend:latest

# CI targets
ci: lint test ## Run CI checks locally (lint + test)
	@echo "CI checks complete"

ci-full: clean deps-audit ci test-coverage ## Full CI pipeline (clean + audit + lint + test + coverage)
	@echo "Full CI pipeline complete"

# Development helpers
dev: ## Run in development mode with auto-reload (requires air)
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/cosmtrek/air@latest; }
	@air

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/cosmtrek/air@latest
	@echo "Tools installed"

#
