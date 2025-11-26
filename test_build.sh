#!/bin/bash

echo "🔨 Testing NetGaze build..."

# Clean everything
rm -rf bin/*
go clean -cache

# Try building with different methods
echo "📦 Method 1: Standard build..."
if go build -o bin/netgaze ./cmd; then
    echo "✅ Method 1 successful!"
    file bin/netgaze
    exit 0
fi

echo "📦 Method 2: Build with tags..."
if go build -ldflags="-s -w" -o bin/netgaze ./cmd; then
    echo "✅ Method 2 successful!"
    file bin/netgaze
    exit 0
fi

echo "📦 Method 3: CGO disabled..."
if CGO_ENABLED=0 go build -o bin/netgaze ./cmd; then
    echo "✅ Method 3 successful!"
    file bin/netgaze
    exit 0
fi

echo "❌ All methods failed!"
exit 1