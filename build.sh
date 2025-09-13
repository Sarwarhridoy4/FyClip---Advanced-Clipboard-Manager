#!/bin/bash

# Create build directory
mkdir -p builds

# Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o builds/fyclip-windows.exe main.go

# macOS
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o builds/fyclip-macos-intel main.go

echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o builds/fyclip-macos-apple main.go

# Linux
echo "Building for Linux (64-bit)..."
GOOS=linux GOARCH=amd64 go build -o builds/fyclip-linux-amd64 main.go

echo "Building for Linux (32-bit)..."
GOOS=linux GOARCH=386 go build -o builds/fyclip-linux-386 main.go

echo "Building for Linux ARM..."
GOOS=linux GOARCH=arm GOARM=7 go build -o builds/fyclip-linux-arm main.go

echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o builds/fyclip-linux-arm64 main.go

# FreeBSD
echo "Building for FreeBSD..."
GOOS=freebsd GOARCH=amd64 go build -o builds/fyclip-freebsd main.go

echo "Build complete! Check the builds/ directory."