#!/bin/bash
# ============================================================
# FyClip - Debian Package Builder
# Converts fyclip.tar.xz (or Go binary) into a .deb installer
# Usage: ./make-deb.sh [version]
# Example: ./make-deb.sh 1.2.0
# ============================================================

APP_NAME="fyclip"
VERSION=${1:-"1.0.0"}   # Default 1.0.0 if no version is passed
ARCH="amd64"
MAINTAINER="Sarwar Hossain <sarwarhridoy4@gmail.com>"
DESCRIPTION="FyClip - Advanced Clipboard Manager
 A cross-platform clipboard manager built with Go and Fyne."

echo "📦 Building ${APP_NAME} version ${VERSION} for ${ARCH}"

# --- Step 1: Clean old build ---
rm -rf ${APP_NAME}-deb ${APP_NAME}_${VERSION}_${ARCH}.deb

# --- Step 2: Extract binary if tar.xz exists ---
if [ -f "${APP_NAME}.tar.xz" ]; then
    echo "📂 Extracting ${APP_NAME}.tar.xz..."
    tar -xf ${APP_NAME}.tar.xz
fi

# --- Step 3: Build binary if missing ---
if [ ! -f "${APP_NAME}" ]; then
    echo "⚙️  Building binary..."
    go build -o ${APP_NAME}
fi

# --- Step 4: Create package structure ---
echo "📂 Creating package structure..."
mkdir -p ${APP_NAME}-deb/DEBIAN
mkdir -p ${APP_NAME}-deb/usr/local/bin
mkdir -p ${APP_NAME}-deb/usr/share/applications
mkdir -p ${APP_NAME}-deb/usr/share/icons/hicolor/64x64/apps

# --- Step 5: Place files ---
echo "📥 Placing files..."
cp ${APP_NAME} ${APP_NAME}-deb/usr/local/bin/
cp icon.png ${APP_NAME}-deb/usr/share/icons/hicolor/64x64/apps/${APP_NAME}.png

# --- Step 6: Control file ---
cat <<EOF > ${APP_NAME}-deb/DEBIAN/control
Package: ${APP_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${MAINTAINER}
Description: ${DESCRIPTION}
EOF

# --- Step 7: Desktop entry ---
cat <<EOF > ${APP_NAME}-deb/usr/share/applications/${APP_NAME}.desktop
[Desktop Entry]
Name=FyClip
Exec=${APP_NAME}
Icon=${APP_NAME}
Type=Application
Categories=Utility;
EOF

# --- Step 8: Build .deb ---
echo "📦 Building .deb package..."
dpkg-deb --build ${APP_NAME}-deb

# --- Step 9: Rename output ---
mv ${APP_NAME}-deb.deb ${APP_NAME}_${VERSION}_${ARCH}.deb

echo "✅ Done! Package created: ${APP_NAME}_${VERSION}_${ARCH}.deb"
echo "👉 Install it with: sudo dpkg -i ${APP_NAME}_${VERSION}_${ARCH}.deb"
