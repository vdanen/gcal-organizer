.PHONY: build test run clean install lint fmt vet

# Binary name
BINARY_NAME=gcal-organizer
BINARY_PATH=./cmd/gcal-organizer

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=gofmt

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run the application
run:
	$(GOCMD) run $(BINARY_PATH) $(ARGS)

# Run with dry-run flag
dry-run:
	$(GOCMD) run $(BINARY_PATH) run --dry-run --verbose

# Install the binary to GOPATH/bin
install:
	$(GOCMD) install $(BINARY_PATH)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Run go vet
vet:
	$(GOVET) ./...

# Format code
fmt:
	$(GOFMT) -w .

# Lint (requires golangci-lint)
lint:
	golangci-lint run

# Check all (format, vet, test)
check: fmt vet test

# Development: build and run
dev: build
	./$(BINARY_NAME) $(ARGS)

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Run the application (use ARGS=... for arguments)"
	@echo "  dry-run       - Run with --dry-run --verbose flags"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  clean         - Remove build artifacts"
	@echo "  vet           - Run go vet"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run golangci-lint"
	@echo "  check         - Run fmt, vet, and test"
	@echo "  dev           - Build and run"
	@echo "  help          - Show this help"
