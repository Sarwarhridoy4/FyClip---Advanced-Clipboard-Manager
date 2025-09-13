#!/bin/bash

# ===========================================================
# Fyne Cross-Platform Build Script
# ===========================================================
# Author: Sarwar Hossain
# Purpose:
#   Automate building Fyne GUI application for Windows, macOS, and Linux.
#   Handles icon bundling, resource generation, cross-compilation,
#   and missing tool installation.
#
# Features:
#   ✅ Fyne icon embedding
#   ✅ Windows ICO & rsrc generation
#   ✅ Cross-platform builds (Win/macOS/Linux)
#   ✅ Automatic tool installation if missing
#   ✅ Detailed logs and stylish output
# ===========================================================

set -e  # Exit immediately if any command fails

# -----------------------------------------------------------
# CONFIGURATION
# -----------------------------------------------------------
ICON_FILE="icon.png"           # Application icon (PNG format)
BUNDLE_DIR="icon"              # Directory for bundled Fyne icon
BUILD_DIR="builds"             # Output directory for binaries
RSRC_FILE="rsrc.syso"          # Windows resource file for icon
MODULE_PATH="fyclip/icon"      # Go module path for bundled icon

# Standard icon sizes for Windows ICO
ICO_SIZES="256,128,64,48,32,16"

# -----------------------------------------------------------
# FUNCTION: Detect system OS and architecture
# -----------------------------------------------------------
detect_system() {
    OS=$(uname -s)
    ARCH=$(uname -m)
    echo "----------------------------------------------------------"
    echo "Detected System:"
    echo "OS          : $OS"
    echo "Architecture: $ARCH"
    echo "----------------------------------------------------------"
}

# -----------------------------------------------------------
# FUNCTION: Ensure required tools are installed
# -----------------------------------------------------------
ensure_tools() {
    echo "Checking required tools..."

    case "$OS" in
        MINGW*|MSYS*|CYGWIN*|Windows_NT)
            # Windows: check ImageMagick & rsrc
            command -v magick >/dev/null 2>&1 || {
                echo "Installing ImageMagick via Chocolatey..."
                choco install imagemagick -y
            }
            command -v rsrc >/dev/null 2>&1 || {
                echo "Installing rsrc via Go..."
                go install github.com/akavel/rsrc@latest
                export PATH=$PATH:$(go env GOPATH)/bin
            }
            ;;
        Linux)
            command -v magick >/dev/null 2>&1 || {
                echo "Installing ImageMagick on Linux..."
                sudo apt-get update
                sudo apt-get install -y imagemagick
            }
            ;;
        Darwin)
            command -v magick >/dev/null 2>&1 || {
                echo "Installing ImageMagick on macOS..."
                brew install imagemagick
            }
            ;;
        *)
            echo "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    echo "All required tools are installed."
    echo "----------------------------------------------------------"
}

# -----------------------------------------------------------
# FUNCTION: Bundle Fyne icon
# Embeds the PNG icon into Go source code
# -----------------------------------------------------------
bundle_icon() {
    echo "Bundling Fyne icon..."
    mkdir -p $BUNDLE_DIR
    fyne bundle $ICON_FILE > $BUNDLE_DIR/icon.go
    echo "Icon successfully bundled."
    echo "----------------------------------------------------------"
}

# -----------------------------------------------------------
# FUNCTION: Build Windows executable with icon
# Steps:
#   1. Convert PNG → ICO
#   2. Generate rsrc.syso for Windows resource
#   3. Build .exe with GUI flag
# -----------------------------------------------------------
build_windows() {
    echo "Building Windows executable..."

    # Convert PNG to ICO
    echo "Generating icon.ico..."
    magick "$ICON_FILE" -define icon:auto-resize=$ICO_SIZES icon.ico

    # Generate Windows resource file
    echo "Generating $RSRC_FILE..."
    rsrc -ico icon.ico -o $RSRC_FILE

    # Build Windows executable
    GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o $BUILD_DIR/fyclip-windows.exe main.go

    # Cleanup temporary files
    rm -f $RSRC_FILE icon.ico

    echo "Windows build complete!"
    echo "----------------------------------------------------------"
}

# -----------------------------------------------------------
# FUNCTION: Build macOS executable
# Supports Intel (amd64) and Apple Silicon (arm64)
# -----------------------------------------------------------
build_macos() {
    echo "Building macOS executables..."

    # Intel
    GOOS=darwin GOARCH=amd64 go build -o $BUILD_DIR/fyclip-macos-intel main.go
    # Apple Silicon
    GOOS=darwin GOARCH=arm64 go build -o $BUILD_DIR/fyclip-macos-apple main.go

    echo "macOS builds complete!"
    echo "----------------------------------------------------------"
}

# -----------------------------------------------------------
# FUNCTION: Build Linux executable
# Supports multiple architectures
# -----------------------------------------------------------
build_linux() {
    echo "Building Linux executables..."

    GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/fyclip-linux-amd64 main.go
    GOOS=linux GOARCH=386 go build -o $BUILD_DIR/fyclip-linux-386 main.go
    GOOS=linux GOARCH=arm GOARM=7 go build -o $BUILD_DIR/fyclip-linux-arm main.go
    GOOS=linux GOARCH=arm64 go build -o $BUILD_DIR/fyclip-linux-arm64 main.go

    echo "Linux builds complete!"
    echo "----------------------------------------------------------"
}

# -----------------------------------------------------------
# FUNCTION: Build FreeBSD executable
# -----------------------------------------------------------
build_freebsd() {
    echo "Building FreeBSD executable..."
    GOOS=freebsd GOARCH=amd64 go build -o $BUILD_DIR/fyclip-freebsd main.go
    echo "FreeBSD build complete!"
    echo "----------------------------------------------------------"
}

# ===========================================================
# MAIN SCRIPT EXECUTION
# ===========================================================

echo "Starting Fyne cross-platform build process..."
detect_system

# Create build output directory if missing
mkdir -p $BUILD_DIR

# Ensure required tools are installed
ensure_tools

# Bundle Fyne icon into Go source
bundle_icon

# Execute builds for all platforms
build_windows
build_macos
build_linux
build_freebsd

echo "=========================================================="
echo "All builds completed successfully!"
echo "Check the '$BUILD_DIR' directory for executables."
echo "=========================================================="
