# NetGaze Makefile

.PHONY: all clean test build run help

# Default target
all: build

# Clean build artifacts
clean:
	rm -f ng
	go clean -cache

# Run tests
test:
	go test ./...

# Build the application
build:
	@echo "Building NetGaze..."
	@if go build -o ng .; then \
		echo "Build successful!"; \
		echo "NetGaze is ready: ./ng"; \
		chmod +x ng; \
		file ng; \
	else \
		echo "Build failed!"; \
		exit 1; \
	fi

# Run the application
run: build
	@echo "Running NetGaze..."
	@./ng --version

# Development mode (uses go run directly)
dev:
	@echo "Development mode (using go run)..."
	@go run main.go --version

# Install to system
install: build
	@echo "Installing NetGaze..."
	@cp ng /usr/local/bin/

# Show help
help:
	@echo "NetGaze Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Build NetGaze binary as 'ng'"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  run       - Run NetGaze (development mode)"
	@echo "  dev       - Development mode (go run)"
	@echo "  install   - Install to /usr/local/bin"
	@echo "  help      - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build the binary as 'ng'"
	@echo "  make run      # Run with built binary"
	@echo "  make dev      # Run with go run (development)"
	@echo "  make test     # Run all tests"
	@echo "  make install   # Install to system"