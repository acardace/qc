.PHONY: build clean install test deps help

# Binary name
BINARY_NAME=qc

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build the binary for Linux
build:
	@echo "Building $(BINARY_NAME) for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -rf reports/

# Install binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Display help
help:
	@echo "Available targets:"
	@echo "  build   - Build the binary for Linux"
	@echo "  clean   - Remove build artifacts"
	@echo "  install - Install binary to GOPATH/bin"
	@echo "  test    - Run tests"
	@echo "  deps    - Download and tidy dependencies"
	@echo "  help    - Display this help message"

# Default target
.DEFAULT_GOAL := build
