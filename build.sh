#!/bin/bash
# =====================================================================
# FyClip Build & Package Script (Debian + AppImage)
# Auto-installs all necessary tools for building
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
APP_NAME_FULL="FyClip - Advanced Clipboard Manager"
BIN_NAME="fyclip"
APP_ID="com.sarwar.fyclip"
PKG_NAME="fyclip"
AUTHOR="Sarwar Hossain"
EMAIL="sarwarhridoy4@gmail.com"
HOMEPAGE="https://github.com/sarwarhridoy/FyClip"
LICENSE="MIT"
CATEGORY="Utility"
SHORT_DESC="Advanced clipboard manager with history, search, and encryption"
LONG_DESC="""\
FyClip is a modular, high-performance clipboard manager built with Go and Fyne.\
 It automatically saves clipboard history including text, images, HTML, and files, with features like:\
  - Clipboard history with persistent storage\
  - Enhanced search with regex, case-sensitive, and fuzzy matching\
  - Encrypted storage with AES-256-GCM\
  - Snippets and templates support\
  - System tray integration\
  - Global hotkey quick access (Ctrl+Shift+V)\
  - Sensitive data detection\n"""
DIST_DIR="dist"
WORK_DIR="${DIST_DIR}/work"
FYNE_ROOT="${WORK_DIR}/fyne-root"
DEB_ROOT="${WORK_DIR}/deb-root"
APPDIR="${WORK_DIR}/FyClip.AppDir"
LOCAL_TOOLS_DIR="$(pwd)/${DIST_DIR}/.tools"

# Detect package manager
detect_package_manager() {
    local pm=""
    for p in apt-get dnf pacman zypper brew; do
        if command -v "${p}" >/dev/null 2>&1; then
            pm="${p}"
            break
        fi
    done
    echo "${pm}"
}

PM=$(detect_package_manager)

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root or have sudo access
has_sudo() {
    if [ "$(id -u)" -eq 0 ]; then return 0; fi
    if command -v sudo >/dev/null 2>&1; then
        if sudo -v 2>/dev/null; then return 0; fi
    fi
    return 1
}

# Install packages based on package manager
install_packages() {
    local packages=("$@")
    local install_cmd=""
    
    case "${PM}" in
        apt-get)
            install_cmd="apt-get update && apt-get install -y ${packages[*]}"
            ;;
        dnf)
            install_cmd="dnf install -y ${packages[*]}"
            ;;
        pacman)
            install_cmd="pacman -Sy --noconfirm ${packages[*]}"
            ;;
        zypper)
            install_cmd="zypper --non-interactive install ${packages[*]}"
            ;;
        brew)
            install_cmd="brew install ${packages[*]}"
            ;;
        *)
            log_error "No supported package manager found!"
            return 1
            ;;
    esac
    
    if has_sudo; then
        if [ "$(id -u)" -eq 0 ]; then
            bash -lc "${install_cmd}"
        else
            sudo bash -lc "${install_cmd}"
        fi
    else
        log_warn "Cannot install packages - no sudo access"
        return 1
    fi
}

# Install Go if not present
install_go() {
    if command -v go >/dev/null 2>&1; then
        log_success "Go is already installed: $(go version)"
        return 0
    fi
    
    log_info "Installing Go..."
    
    local go_version="1.21.5"
    local arch
    arch=$(uname -m)
    case "${arch}" in
        x86_64) arch="amd64" ;;
        aarch64) arch="arm64" ;;
        armv7l) arch="armv6l" ;;
    esac
    
    local go_archive="go${go_version}.linux-${arch}.tar.gz"
    local tmp_dir=$(mktemp -d)
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://go.dev/dl/${go_archive}" -o "${tmp_dir}/${go_archive}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "https://go.dev/dl/${go_archive}" -O "${tmp_dir}/${go_archive}"
    else
        log_error "Neither curl nor wget found. Cannot download Go."
        return 1
    fi
    
    tar -C "${tmp_dir}" -xzf "${tmp_dir}/${go_archive}"
    
    if [ "$(id -u)" -eq 0 ]; then
        mv "${tmp_dir}/go" /usr/local/go
    else
        mkdir -p "${HOME}/.local/go"
        mv "${tmp_dir}/go" "${HOME}/.local/go"
    fi
    
    rm -rf "${tmp_dir}"
    
    # Add to PATH
    if [ "$(id -u)" -eq 0 ]; then
        echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/golang.sh
    fi
    export PATH="${HOME}/.local/go/bin:${PATH}"
    
    log_success "Go installed successfully"
    return 0
}

# Install fyne CLI
install_fyne() {
    if command -v fyne >/dev/null 2>&1; then
        log_success "fyne CLI is already installed"
        return 0
    fi
    
    log_info "Installing fyne CLI..."
    
    # Ensure Go is installed first
    if ! command -v go >/dev/null 2>&1; then
        install_go || return 1
    fi
    
    export PATH="${HOME}/.local/go/bin:${PATH}:$(go env GOPATH)/bin"
    
    if go install fyne.io/tools/cmd/fyne@latest; then
        log_success "fyne CLI installed successfully"
        return 0
    else
        log_error "Failed to install fyne CLI"
        return 1
    fi
}

# Install appimagetool
install_appimagetool() {
    if command -v appimagetool >/dev/null 2>&1; then
        log_success "appimagetool is already installed"
        return 0
    fi
    
    log_info "Installing appimagetool..."
    
    local arch="x86_64"
    case $(uname -m) in
        aarch64) arch="aarch64" ;;
        armv7l) arch="armhf" ;;
    esac
    
    local url="https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${arch}.AppImage"
    local tmp_file=$(mktemp)
    
    mkdir -p "${LOCAL_TOOLS_DIR}/bin"
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "${url}" -o "${tmp_file}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "${url}" -O "${tmp_file}"
    else
        log_error "Neither curl nor wget found"
        return 1
    fi
    
    mv "${tmp_file}" "${LOCAL_TOOLS_DIR}/bin/appimagetool"
    chmod +x "${LOCAL_TOOLS_DIR}/bin/appimagetool"
    
    log_success "appimagetool installed successfully"
    return 0
}

# Install required system packages
install_system_dependencies() {
    log_info "Installing system dependencies..."
    
    local packages=()
    
    case "${PM}" in
        apt-get)
            packages=(build-essential curl wget tar dpkg libgl1-mesa-glx libglib2.0-0 libxcursor1 libxrandr2 libxinerama1 libxi6 libgtk-3-0 xclip xsel wl-clipboard)
            ;;
        dnf)
            packages=( @development-tools curl wget tar dpkg glibc libXcursor libXrandr libXinerama libXi gtk3 xclip xsel wl-clipboard)
            ;;
        pacman)
            packages=(base-devel curl wget tar dpkg libglvnd libxcursor libxrandr libxinerama libxi gtk3 xclip xsel wl-clipboard)
            ;;
        zypper)
            packages=(-devel_basis curl wget tar dpkg Mesa-libGL1 glib2 libxcursor1 libXrandr2 libXinerama1 libXi6 gtk3-tools xclip xsel wl-clipboard)
            ;;
        brew)
            packages=(curl wget tar dpkg)
            ;;
    esac
    
    if [ ${#packages[@]} -gt 0 ]; then
        install_packages "${packages[@]}" || log_warn "Some packages may have failed to install"
    fi
    
    log_success "System dependencies installed"
    return 0
}

# Ensure all required tools are available
ensure_tools() {
    log_info "Checking required tools..."
    
    # Install system dependencies first
    install_system_dependencies || true
    
    # Check and install Go
    if ! command -v go >/dev/null 2>&1; then
        install_go || { log_error "Failed to install Go"; exit 1; }
    fi
    export PATH="${HOME}/.local/go/bin:${PATH}:$(go env GOPATH 2>/dev/null)/bin"
    log_success "Go: $(go version)"
    
    # Check and install fyne
    if ! command -v fyne >/dev/null 2>&1; then
        install_fyne || { log_error "Failed to install fyne CLI"; exit 1; }
    fi
    log_success "fyne CLI installed"
    
    # Check for tar
    if ! command -v tar >/dev/null 2>&1; then
        install_packages tar || { log_error "Failed to install tar"; exit 1; }
    fi
    
    # Check for dpkg-deb
    if ! command -v dpkg-deb >/dev/null 2>&1; then
        install_packages dpkg || { log_error "Failed to install dpkg"; exit 1; }
    fi
    
    # Check and install appimagetool
    if ! command -v appimagetool >/dev/null 2>&1; then
        install_appimagetool || { log_error "Failed to install appimagetool"; exit 1; }
    fi
    
    # Add local tools to PATH
    export PATH="${LOCAL_TOOLS_DIR}/bin:${PATH}"
    
    log_success "All required tools are available"
}

# Get version
get_version() {
    if git rev-parse --git-dir >/dev/null 2>&1; then
        git describe --tags --abbrev=0 2>/dev/null || echo "1.0.0"
    else
        echo "1.0.0"
    fi
}

# Main build function
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  FyClip Build Script v2.1.2${NC}"
    echo -e "${BLUE}  Auto-installs all dependencies${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # Get version
    if [ -n "${1:-}" ]; then 
        VERSION="$1"
    else
        DEFAULT_VERSION=$(get_version)
        read -r -p "Enter version [${DEFAULT_VERSION}]: " VERSION
        VERSION=${VERSION:-$DEFAULT_VERSION}
    fi
    
    # Detect architecture
    ARCH_RAW=$(uname -m)
    case "$ARCH_RAW" in
        x86_64) ARCH="amd64"; APPIMAGE_ARCH="x86_64" ;;
        aarch64) ARCH="arm64"; APPIMAGE_ARCH="aarch64" ;;
        armv7l) ARCH="armhf"; APPIMAGE_ARCH="armhf" ;;
        i386|i686) ARCH="i386"; APPIMAGE_ARCH="i686" ;;
        *) ARCH="$ARCH_RAW"; APPIMAGE_ARCH="$ARCH_RAW" ;;
    esac
    
    log_info "Building ${APP_NAME} ${VERSION} (${ARCH})"
    
    # Ensure all tools are available
    ensure_tools
    
    # Create directories
    mkdir -p "${DIST_DIR}" "${WORK_DIR}"
    rm -rf "${FYNE_ROOT}" "${DEB_ROOT}" "${APPDIR}"
    rm -f "${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb" \
           "${DIST_DIR}/${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage" \
           "${APP_NAME}.tar.xz"
    
    # Build with Fyne
    log_info "Building FyClip with Fyne..."
    fyne package --os linux --release --name "${APP_NAME}" --icon icon.png
    
    # Extract Fyne package
    log_info "Extracting Fyne package..."
    tar -xJf "${APP_NAME}.tar.xz" -C "${WORK_DIR}"
    
    # Detect prefix
    if [ -d "${WORK_DIR}/usr/local" ]; then 
        PREFIX_REL="usr/local"
    else 
        PREFIX_REL="usr"
    fi
    
    # Normalize directory structure
    USR_NORMALIZED="${WORK_DIR}/usr-normalized"
    mkdir -p "${USR_NORMALIZED}"
    cp -a "${WORK_DIR}/${PREFIX_REL}/." "${USR_NORMALIZED}/"
    
    BIN_DIR="${USR_NORMALIZED}/bin"
    APPS_DIR="${USR_NORMALIZED}/share/applications"
    PIXMAPS_DIR="${USR_NORMALIZED}/share/pixmaps"
    
    # Normalize binary name (only if different)
    FOUND_BIN=$(find "${BIN_DIR}" -maxdepth 1 -type f -executable | head -n 1)
    if [ -z "${FOUND_BIN}" ]; then
        log_error "Binary not found"
        exit 1
    fi
    BIN_BASENAME=$(basename "${FOUND_BIN}")
    if [ "${BIN_BASENAME}" != "${BIN_NAME}" ]; then
        mv "${FOUND_BIN}" "${BIN_DIR}/${BIN_NAME}"
        BIN_PATH="${BIN_DIR}/${BIN_NAME}"
    else
        BIN_PATH="${FOUND_BIN}"
    fi
    
    # Find desktop and icon files
    DESKTOP_PATH=$(find "${APPS_DIR}" -name '*.desktop' | head -n 1)
    ICON_PATH=$(find "${PIXMAPS_DIR}" -type f | head -n 1)
    
    if [ -z "${DESKTOP_PATH}" ]; then
        log_error "Desktop file not found"
        exit 1
    fi
    if [ -z "${ICON_PATH}" ]; then
        log_error "Icon file not found"
        exit 1
    fi
    
    # Normalize desktop and icon filenames
    DESKTOP_DIR="$(dirname "${DESKTOP_PATH}")"
    DESKTOP_NORM="${DESKTOP_DIR}/${APP_ID}.desktop"
    if [ "$(basename "${DESKTOP_PATH}")" != "${APP_ID}.desktop" ]; then
        mv "${DESKTOP_PATH}" "${DESKTOP_NORM}"
        DESKTOP_PATH="${DESKTOP_NORM}"
    fi
    
    ICON_DIR="$(dirname "${ICON_PATH}")"
    ICON_EXT="${ICON_PATH##*.}"
    ICON_NORM="${ICON_DIR}/${APP_ID}.${ICON_EXT}"
    if [ "$(basename "${ICON_PATH}")" != "${APP_ID}.${ICON_EXT}" ]; then
        mv "${ICON_PATH}" "${ICON_NORM}"
        ICON_PATH="${ICON_NORM}"
    fi
    
    # Update desktop file with proper Debian/AppImage metadata
    sed -i -E "s|^Exec=.*|Exec=${BIN_NAME}|g" "${DESKTOP_PATH}"
    sed -i -E "s|^Icon=.*|Icon=${APP_ID}|g" "${DESKTOP_PATH}"
    sed -i -E "s|^Name=.*|Name=${APP_NAME}|g" "${DESKTOP_PATH}"
    
    # Ensure proper desktop file fields per Debian standards
    sed -i -E "s|^Categories=.*|Categories=${CATEGORY};|g" "${DESKTOP_PATH}"
    grep -q '^Categories=' "${DESKTOP_PATH}" || echo "Categories=${CATEGORY};" >> "${DESKTOP_PATH}"
    grep -q '^NoDisplay=' "${DESKTOP_PATH}" || echo "NoDisplay=false" >> "${DESKTOP_PATH}"
    grep -q '^Keywords=' "${DESKTOP_PATH}" || echo "Keywords=clipboard;copy;paste;history;manager;" >> "${DESKTOP_PATH}"
    grep -q '^StartupWMClass=' "${DESKTOP_PATH}" || echo "StartupWMClass=${APP_NAME}" >> "${DESKTOP_PATH}"
    
    # Add Terminal=false for desktop environments
    grep -q '^Terminal=' "${DESKTOP_PATH}" || echo "Terminal=false" >> "${DESKTOP_PATH}"
    
    # Add MIME types for clipboard content
    grep -q '^MimeType=' "${DESKTOP_PATH}" || echo "MimeType=text/plain;text/html;image/png;image/jpeg;application/octet-stream;" >> "${DESKTOP_PATH}"
    
    # Add generic name
    grep -q '^GenericName=' "${DESKTOP_PATH}" || echo "GenericName=Clipboard Manager" >> "${DESKTOP_PATH}"
    
    # Add screenshot field for software center (Debian)
    ScreenshotPath="${USR_NORMALIZED}/share/app/screenshots/screenshot1.png"
    if [ -f "internal/app/assets/screenshots/screenshot1.png" ]; then
        mkdir -p "${USR_NORMALIZED}/share/app/screenshots"
        cp -f "internal/app/assets/screenshots/screenshot1.png" "${USR_NORMALIZED}/share/app/screenshots/"
        grep -q '^X-AppImage-Comment=' "${DESKTOP_PATH}" || echo "X-AppImage-Comment=Generated by FyClip" >> "${DESKTOP_PATH}"
    fi
    
    # Install icon in hicolor
    HICOLOR_APPS_DIR="${USR_NORMALIZED}/share/icons/hicolor/256x256/apps"
    mkdir -p "${HICOLOR_APPS_DIR}"
    cp -f "${ICON_PATH}" "${HICOLOR_APPS_DIR}/${APP_ID}.${ICON_EXT}"
    
    # ---------------------------------------------------------------------
    # Build Debian Package
    # ---------------------------------------------------------------------
    log_info "Building Debian package..."
    mkdir -p "${DEB_ROOT}/DEBIAN" "${DEB_ROOT}/usr"
    cp -a "${USR_NORMALIZED}/." "${DEB_ROOT}/usr/"
    
    # Create proper DEBIAN/control file with full metadata per Debian policy
    cat > "${DEB_ROOT}/DEBIAN/control" <<EOF
Package: ${PKG_NAME}
Version: ${VERSION}
Section: ${CATEGORY}
Priority: optional
Architecture: ${ARCH}
Depends: ${PKG_NAME}-common (>= ${VERSION}), \
         libgl1, \
         libx11-6, \
         libxcursor1, \
         libxrandr2, \
         libxinerama1, \
         libxi6, \
         libgtk-3-0, \
         xclip | xsel, \
         wl-clipboard
Maintainer: ${AUTHOR} <${EMAIL}>
Description: ${SHORT_DESC}
 ${LONG_DESC}
EOF
    
    # Create DEBIAN/md5sums for package integrity
    mkdir -p "${DEB_ROOT}/DEBIAN"
    
    # Create postinst script
    cat > "${DEB_ROOT}/DEBIAN/postinst" <<'EOF'
#!/bin/sh
# postinst script for fyclip
set -e

case "$1" in
    configure)
        # Update desktop database
        if command -v update-desktop-database >/dev/null 2>&1; then
            update-desktop-database -q /usr/share/applications 2>/dev/null || true
        fi
        
        # Update icon cache
        if command -v gtk-update-icon-cache >/dev/null 2>&1; then
            gtk-update-icon-cache -q -t /usr/share/icons/hicolor 2>/dev/null || true
        fi
        ;;
    abort-upgrade|abort-remove|abort-install)
        ;;
    *)
        ;;
esac

# Enable autostart if requested
#if [ -d "$HOME/.config/autostart" ]; then
#    cp /usr/share/applications/${APP_ID}.desktop "$HOME/.config/autostart/" 2>/dev/null || true
#fi

exit 0
EOF
    chmod 755 "${DEB_ROOT}/DEBIAN/postinst"
    
    # Create prerm script for cleanup
    cat > "${DEB_ROOT}/DEBIAN/prerm" <<'EOF'
#!/bin/sh
# prerm script for fyclip
set -e

case "$1" in
    remove|deconfigure)
        # Remove desktop database entry
        if command -v update-desktop-database >/dev/null 2>&1; then
            update-desktop-database -q /usr/share/applications 2>/dev/null || true
        fi
        ;;
    upgrade|failed-upgrade)
        ;;
    *)
        ;;
esac

exit 0
EOF
    chmod 755 "${DEB_ROOT}/DEBIAN/prerm"
    
    # Create copyright file (required by Debian policy)
    mkdir -p "${DEB_ROOT}/usr/share/doc/${PKG_NAME}"
    cat > "${DEB_ROOT}/usr/share/doc/${PKG_NAME}/copyright" <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: ${APP_NAME}
Upstream-Contact: ${AUTHOR} <${EMAIL}>
Source: ${HOMEPAGE}

Files: *
Copyright: $(date +%Y) ${AUTHOR}
License: ${LICENSE}
 ${LONG_DESC}
EOF
    gzip -n "${DEB_ROOT}/usr/share/doc/${PKG_NAME}/copyright"
    
    # Create changelog file
    mkdir -p "${DEB_ROOT}/usr/share/doc/${PKG_NAME}"
    cat > "${DEB_ROOT}/usr/share/doc/${PKG_NAME}/changelog" <<EOF
${PKG_NAME} (${VERSION}) stable; urgency=low

  * Initial release

 -- ${AUTHOR} <${EMAIL}>  $(date -R)
EOF
    gzip -n "${DEB_ROOT}/usr/share/doc/${PKG_NAME}/changelog"
    
    # Create man page placeholder
    mkdir -p "${DEB_ROOT}/usr/share/man/man1"
    cat > "${DEB_ROOT}/usr/share/man/man1/${PKG_NAME}.1" <<EOF
.TH ${PKG_NAME^^} 1 "$(date +%Y)"
.SH NAME
${PKG_NAME} \- ${SHORT_DESC}
.SH SYNOPSIS
.B ${PKG_NAME}
.RI [ options ]
.SH DESCRIPTION
${LONG_DESC}
.SH OPTIONS
.TP
.B \-h, \-help
Show help message
.SH AUTHOR
Written by ${AUTHOR} <${EMAIL}>
.SH REPORTING BUGS
Report bugs at ${HOMEPAGE}/issues
.SH SEE ALSO
.BR xclip (1), BR xsel (1)
EOF
    gzip -n "${DEB_ROOT}/usr/share/man/man1/${PKG_NAME}.1"
    
    # Remove existing package file if present
    rm -f "${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb"
    
    # Build the package with md5sums
    (cd "${DEB_ROOT}" && find usr -type f -exec md5sum {} \; > DEBIAN/md5sums)
    dpkg-deb --build "${DEB_ROOT}" "${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb"
    log_success "Debian package built: ${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb"
    
    # Remove previously installed package to avoid conflicts
    echo "Removing previously installed package..."
    dpkg -r "${PKG_NAME}" 2>/dev/null || true
    
    # ---------------------------------------------------------------------
    # Build AppImage
    # ---------------------------------------------------------------------
    log_info "Building AppImage..."
    mkdir -p "${APPDIR}/usr"
    cp -a "${USR_NORMALIZED}/." "${APPDIR}/usr/"
    
    # Create AppRun with proper AppImage metadata
    cat > "${APPDIR}/AppRun" <<'EOF'
#!/bin/sh
# AppRun script for FyClip AppImage
# This script is required for AppImage to work properly

# Get the directory where this script is located
APPIMAGE="$(readlink -f "${0}")"
HERE="$(dirname "${APPIMAGE}")"

# Export library path
export LD_LIBRARY_PATH="${HERE}/usr/lib:${HERE}/usr/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH}"
export XDG_DATA_DIRS="${HERE}/usr/share:${XDG_DATA_DIRS}"
export XDG_CONFIG_DIRS="${HERE}/etc/xdg:${XDG_CONFIG_DIRS}"

# Set QT plugin path for Fyne
export QT_PLUGIN_PATH="${HERE}/usr/plugins:${QT_PLUGIN_PATH}"
export QML2_IMPORT_PATH="${HERE}/usr/qml:${QML2_IMPORT_PATH}"

# Execute the application
exec "${HERE}/usr/bin/${BIN_NAME}" "$@"
EOF
    chmod +x "${APPDIR}/AppRun"
    
    # Copy desktop file with proper AppImage metadata
    cp "${DESKTOP_PATH}" "${APPDIR}/${APP_ID}.desktop"
    
    # Update desktop file for AppImage with all required fields
    sed -i -E "s|^Exec=.*|Exec=AppRun|g" "${APPDIR}/${APP_ID}.desktop"
    sed -i -E "s|^Icon=.*|Icon=${APP_ID}|g" "${APPDIR}/${APP_ID}.desktop"
    sed -i -E "s|^Path=.*||g" "${APPDIR}/${APP_ID}.desktop"
    
    # Add AppImage-specific desktop file fields
    grep -q '^TryExec=' "${APPDIR}/${APP_ID}.desktop" || echo "TryExec=AppRun" >> "${APPDIR}/${APP_ID}.desktop"
    
    # Copy icon to multiple locations for AppImage
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/256x256/apps"
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/128x128/apps"
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/64x64/apps"
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/48x48/apps"
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/32x32/apps"
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/22x22/apps"
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/16x16/apps"
    
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/256x256/apps/${APP_ID}.${ICON_EXT}"
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/128x128/apps/${APP_ID}.${ICON_EXT}"
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/64x64/apps/${APP_ID}.${ICON_EXT}"
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/48x48/apps/${APP_ID}.${ICON_EXT}"
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/32x32/apps/${APP_ID}.${ICON_EXT}"
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/22x22/apps/${APP_ID}.${ICON_EXT}"
    cp "${ICON_PATH}" "${APPDIR}/usr/share/icons/hicolor/16x16/apps/${APP_ID}.${ICON_EXT}"
    
    # Copy icon to root of AppDir
    cp "${ICON_PATH}" "${APPDIR}/${APP_ID}.${ICON_EXT}"
    
    # Create .DirIcon symlink (required for AppImage)
    ln -sf "${APP_ID}.${ICON_EXT}" "${APPDIR}/.DirIcon"
    
    # Copy screenshots for AppImage (if available)
    mkdir -p "${APPDIR}/usr/share/app/screenshots"
    if [ -d "internal/app/assets/screenshots" ]; then
        cp internal/app/assets/screenshots/*.png "${APPDIR}/usr/share/app/screenshots/" 2>/dev/null || true
    fi
    
    # Create AppImage metadata file (appimagetool uses this)
    cat > "${APPDIR}/appimagetool.yaml" <<'EOF'
appId: ${APP_ID}
productName: ${APP_NAME}
version: ${VERSION}
description: ${SHORT_DESC}
author: ${AUTHOR} <${EMAIL}>
license: ${LICENSE}
categories:
  - ${CATEGORY}
homepage: ${HOMEPAGE}

# Icon (will use .DirIcon if not specified)
icon: ${APP_ID}.${ICON_EXT}

# Desktop file
desktopFile: ${APP_ID}.desktop
EOF
    
    # Create a metainfo file for AppImage (AppStream standard)
    mkdir -p "${APPDIR}/usr/share/metainfo"
    cat > "${APPDIR}/usr/share/metainfo/${APP_ID}.appdata.xml" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>${APP_ID}</id>
  <name>${APP_NAME}</name>
  <summary>${SHORT_DESC}</summary>
  <metadata_license>CC0-1.0</metadata_license>
  <project_license>${LICENSE}</project_license>
  <description>
    <p>${LONG_DESC}</p>
  </description>
  <categories>
    <category>${CATEGORY}</category>
  </categories>
  <url type="homepage">${HOMEPAGE}</url>
  <launchable type="desktop-id">${APP_ID}.desktop</launchable>
  <provides>
    <binary>${BIN_NAME}</binary>
  </provides>
  <releases>
    <release version="${VERSION}" date="$(date +%Y-%m-%d)">
      <description>
        <p>Initial release</p>
      </description>
    </release>
  </releases>
  <content_rating type="oars-1.1" />
</component>
EOF
    
    # Copy metainfo to AppDir root for AppImage
    cp "${APPDIR}/usr/share/metainfo/${APP_ID}.appdata.xml" "${APPDIR}/"
    
    # Create .desktop file in AppDir root (for some desktop environments)
    cp "${APPDIR}/${APP_ID}.desktop" "${APPDIR}/fyclip.desktop"
    
    # Bundle dependencies
    mkdir -p "${APPDIR}/usr/lib"
    ldd "${BIN_PATH}" 2>/dev/null | awk '/=> \// {print $3}' | \
        grep -vE 'lib(c|pthread|rt|dl|m|gcc_s|stdc\+\+|glib|X11|xcb|wayland)' | \
        xargs -r -I '{}' cp '{}' "${APPDIR}/usr/lib/" 2>/dev/null || true
    
    # Build AppImage
    # Use absolute path since we cd to DIST_DIR
    APPDIR_ABS="$(pwd)/${WORK_DIR}"
    cd "${DIST_DIR}"
    APPIMAGE_EXTRACT_AND_RUN=1 "${LOCAL_TOOLS_DIR}/bin/appimagetool" \
        "${APPDIR_ABS}/FyClip.AppDir" \
        "${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"
    cd - >/dev/null
    
    log_success "AppImage built: ${DIST_DIR}/${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"
    
    # Save the original Fyne tar.xz for distribution
    if [ -f "${APP_NAME}.tar.xz" ]; then
        cp "${APP_NAME}.tar.xz" "${DIST_DIR}/${PKG_NAME}_${VERSION}_linux_${APPIMAGE_ARCH}.tar.xz"
        log_success "Source tarball saved: ${DIST_DIR}/${PKG_NAME}_${VERSION}_linux_${APPIMAGE_ARCH}.tar.xz"
    fi
    
    # Cleanup - keep only dist folder contents
    rm -rf "${WORK_DIR}"
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Build Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "Debian:   ${BLUE}${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb${NC}"
    echo -e "AppImage: ${BLUE}${DIST_DIR}/${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage${NC}"
    echo -e "Tarball: ${BLUE}${DIST_DIR}/${PKG_NAME}_${VERSION}_linux_${APPIMAGE_ARCH}.tar.xz${NC}"
    echo ""
}

# Run main function
main "$@"
