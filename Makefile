.PHONY: build test clean install dev lint coverage help

# Binary name
BINARY_NAME=migra
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

help: ## Display this help message
	@echo "Migra - Migration Orchestration CLI"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/migra/main.go

dev: ## Build and run with example config
	@echo "Building and running $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/migra/main.go
	./bin/$(BINARY_NAME) version

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/migra

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/ dist/ coverage.txt coverage.html
	go clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

snapshot: ## Build snapshot with GoReleaser
	@echo "Building snapshot..."
	goreleaser release --snapshot --clean

.DEFAULT_GOAL := help
