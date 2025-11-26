#!/bin/bash

echo "Building NetGaze..."

# Clean everything
rm -rf bin/*
go clean -cache

# Try building with different methods
echo "Method 1: Standard build..."
if go build -o bin/netgaze .; then
    echo "Method 1 successful!"
    file bin/netgaze
    exit 0
fi

echo "Method 2: Build with tags..."
if go build -ldflags="-s -w" -o bin/netgaze .; then
    echo "Method 2 successful!"
    file bin/netgaze
    exit 0
fi

echo "Method 3: CGO disabled..."
if CGO_ENABLED=0 go build -o bin/netgaze .; then
    echo "Method 3 successful!"
    file bin/netgaze
    exit 0
fi

echo "All methods failed!"
exit 1

help:
	@echo "NetGaze Build Script"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Build NetGaze binary"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  run       - Run NetGaze (development mode)"
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build the binary"
	@echo "  make run      # Run with go run (development)"
	@echo "  make test     # Run all tests"