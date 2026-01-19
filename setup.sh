#!/bin/bash

echo "=== DNS DDoS Defense - Setup Script ==="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    echo "   Visit: https://go.dev/dl/"
    exit 1
fi

echo "✓ Go found: $(go version)"
echo ""

# Create necessary directories
echo "Creating directories..."
mkdir -p logs
mkdir -p configs
echo "✓ Directories created"
echo ""

# Download dependencies
echo "Downloading dependencies..."
go mod download
if [ $? -ne 0 ]; then
    echo "❌ Failed to download dependencies"
    exit 1
fi

echo "✓ Dependencies downloaded"
echo ""

# Tidy up modules
echo "Tidying up modules..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "❌ Failed to tidy modules"
    exit 1
fi

echo "✓ Modules tidied"
echo ""

# Build the application
echo "Building application..."
go build -o dns-defense-server ./cmd/server
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✓ Build successful"
echo ""

echo "=== Setup Complete! ==="
echo ""
echo "Quick Start:"
echo "  1. Run on port 5353 (no sudo): ./dns-defense-server -port 5353"
echo "  2. Run on port 53 (requires sudo): sudo ./dns-defense-server"
echo "  3. Test: dig @localhost -p 5353 google.com"
echo ""
echo "For more options, see README.md or QUICKSTART.md"
