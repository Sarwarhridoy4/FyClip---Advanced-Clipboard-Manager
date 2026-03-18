# FyClip Makefile
# Build and release automation

# Variables
APP_NAME := FyClip
BINARY_NAME := fyclip
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Go parameters
GO := go
GOCMD := $(GO)
GOBUILD := $(GOCMD) build $(LDFLAGS)
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOFMT := $(GOCMD) fmt
GOLINT := $(GOCMD) lint

# Directories
OUTPUT_DIR := bin
SOURCE_DIRS := .

# Platform-specific settings
ifeq ($(OS),Windows_NT)
	EXT := .exe
else
	EXT :=
endif

# Default target
.PHONY: all
all: lint test build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)$(EXT) .

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(OUTPUT_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/linux/$(BINARY_NAME) .
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(OUTPUT_DIR)/linux/$(BINARY_NAME)-arm64 .

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(OUTPUT_DIR)/darwin
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/darwin/$(BINARY_NAME) .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(OUTPUT_DIR)/darwin/$(BINARY_NAME)-arm64 .

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(OUTPUT_DIR)/windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/windows/$(BINARY_NAME).exe .

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOVET) ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w $(SOURCE_DIRS)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(OUTPUT_DIR)
	rm -f coverage.out

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Run the application
.PHONY: run
run:
	$(GOCMD) run .

# Generate coverage report
.PHONY: coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Build with Fyne packaging
.PHONY: package
package: build
	@echo "Packaging application..."
	./build.sh

# Create release
.PHONY: release
release: clean test lint build-all package
	@echo "Creating release..."
	@mkdir -p release
	cd $(OUTPUT_DIR) && zip -r ../release/$(APP_NAME)-$(VERSION)-all.zip ./

# Development shortcuts
.PHONY: dev
dev: deps run

.PHONY: check
check: lint test

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Run lint, test and build (default)"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for Linux, Darwin, and Windows"
	@echo "  test         - Run tests"
	@echo "  benchmark    - Run benchmarks"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  run          - Run the application"
	@echo "  coverage     - Generate coverage report"
	@echo "  package      - Package the application"
	@echo "  release      - Create release"
	@echo "  dev          - Development mode"
	@echo "  check        - Run lint and tests"
	@echo "  help         - Show this help message"
