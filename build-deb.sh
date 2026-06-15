#!/bin/bash
# =====================================================================
# FyClip Debian Package Build Script
# =====================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="FyClip"
BIN_NAME="fyclip"
APP_ID="com.sarwar.fyclip"
PKG_NAME="fyclip"
AUTHOR="Sarwar Hossain"
EMAIL="sarwarhridoy4@gmail.com"
DIST_DIR="dist"
DEB_BUILD_DIR="${DIST_DIR}/deb-build"

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Get version from git or use default
get_version() {
    if git rev-parse --git-dir >/dev/null 2>&1; then
        git fetch --tags 2>/dev/null || true
        git tag --sort=-v:refname | head -1 | sed 's/^v//' || echo "2.2.2"
    else
        echo "2.2.2"
    fi
}

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing=()
    
    if ! command -v go >/dev/null 2>&1; then
        missing+=("golang-go")
    fi
    
    if ! command -v dpkg-buildpackage >/dev/null 2>&1; then
        missing+=("dpkg-dev")
    fi
    
    if ! command -v dh >/dev/null 2>&1; then
        missing+=("debhelper")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing[*]}"
        log_info "Install them with:"
        log_info "  sudo apt-get install ${missing[*]}"
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."
    rm -rf "${DEB_BUILD_DIR}"
    rm -f "${DIST_DIR}"/*.deb
    rm -f "${DIST_DIR}"/*.build
    rm -f "${DIST_DIR}"/*.buildinfo
    rm -f "${DIST_DIR}"/*.changes
    rm -f "${DIST_DIR}"/*.dsc
    rm -f "${DIST_DIR}"/*.tar.gz
    rm -f "${DIST_DIR}"/*.tar.xz
    log_success "Cleaned previous builds"
}

# Prepare source for Debian build
prepare_source() {
    log_info "Preparing source for Debian build..."
    
    mkdir -p "${DEB_BUILD_DIR}"
    
    # Copy source files (excluding dist directory to avoid copying into itself)
    rsync -a --exclude='dist' --exclude='.git' --exclude='bin' --exclude='snap' --exclude='.github' . "${DEB_BUILD_DIR}/"
    
    # Create tarball
    cd "${DEB_BUILD_DIR}"
    tar -czf "../${PKG_NAME}_${VERSION}.orig.tar.gz" --exclude='debian' .
    cd - > /dev/null
    
    log_success "Source prepared"
}

# Build Debian package
build_deb_package() {
    log_info "Building Debian package..."
    
    cd "${DEB_BUILD_DIR}"
    
    # Build the package
    dpkg-buildpackage -us -uc -b
    
    cd - > /dev/null
    
    # Move .deb file to dist directory
    mv "${DIST_DIR}/../${PKG_NAME}_${VERSION}"*.deb "${DIST_DIR}/" 2>/dev/null || true
    
    log_success "Debian package built successfully"
}

# Verify the package
verify_package() {
    log_info "Verifying package..."
    
    local deb_file=$(find "${DIST_DIR}" -name "${PKG_NAME}_${VERSION}-*.deb" | head -n 1)
    
    if [ -z "${deb_file}" ]; then
        log_error "Package file not found"
        return 1
    fi
    
    log_info "Package: ${deb_file}"
    
    # Show package info
    dpkg-deb --info "${deb_file}"
    
    # Show package contents
    log_info "Package contents:"
    dpkg-deb --contents "${deb_file}"
    
    log_success "Package verified"
}

# Main function
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  FyClip Debian Package Builder${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # Get version
    if [ -n "${1:-}" ]; then
        VERSION="$1"
    else
        VERSION=$(get_version)
    fi
    
    log_info "Building ${APP_NAME} ${VERSION} Debian package"
    
    # Check dependencies
    check_dependencies
    
    # Create dist directory
    mkdir -p "${DIST_DIR}"
    
    # Clean previous builds
    clean_build
    
    # Prepare source
    prepare_source
    
    # Build package
    build_deb_package
    
    # Verify package
    verify_package
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Build Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    log_success "Debian package is available in ${DIST_DIR}/"
    echo ""
    log_info "To install the package:"
    log_info "  sudo dpkg -i ${DIST_DIR}/${PKG_NAME}_${VERSION}*.deb"
    log_info "  sudo apt-get install -f  # Fix any dependency issues"
    echo ""
    log_info "To remove the package:"
    log_info "  sudo dpkg -r ${PKG_NAME}"
}

# Run main function
main "$@"
