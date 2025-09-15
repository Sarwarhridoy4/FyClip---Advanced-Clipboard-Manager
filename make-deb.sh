#!/bin/bash
# ============================================================
# FyClip Build & Package Script (Fully Automated)
# Usage: ./build-all.sh [version]
# Ensures tray icon works by including icon.png and setting AppID
# ============================================================

APP_NAME="fyclip"
VERSION=${1:-"1.0.0"}
APP_ID="com.sarwar.fyclip"   # Unique AppID used by Fyne

# --- Map uname -m to Debian and AppImage architectures ---
ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64) ARCH="amd64"; APPIMAGE_ARCH="x86_64" ;;
  aarch64) ARCH="arm64"; APPIMAGE_ARCH="arm64" ;;
  armv7l) ARCH="armhf"; APPIMAGE_ARCH="armhf" ;;
  i386) ARCH="i386"; APPIMAGE_ARCH="i386" ;;
  *) ARCH="$ARCH_RAW"; APPIMAGE_ARCH="$ARCH_RAW" ;;
esac

AUTHOR="Sarwar Hossain"
EMAIL="sarwarhridoy4@gmail.com"
REPO="https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager"
DESCRIPTION_SHORT="FyClip - Advanced Clipboard Manager"
DESCRIPTION_LONG="A powerful, cross-platform clipboard manager built with Go and Fyne that automatically tracks your clipboard history, provides instant search, and persists data between sessions. Now with image support and pinning for favorite items."

echo "📦 Building ${APP_NAME} version ${VERSION} for ${ARCH}"

# ============================================================
# Step 0: Install necessary tools
# ============================================================
echo "🔧 Installing required tools (ImageMagick, wget, dpkg-deb, Go)..."
sudo apt-get update
sudo apt-get install -y imagemagick wget dpkg-dev golang-go

# ============================================================
# Step 1: Clean previous builds
# ============================================================
rm -rf ${APP_NAME}-deb ${APP_NAME}_${VERSION}_${ARCH}.deb FyClip.AppDir ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage

# ============================================================
# Step 2: Extract binary if tar.xz exists
# ============================================================
if [ -f "${APP_NAME}.tar.xz" ]; then
    echo "📂 Extracting ${APP_NAME}.tar.xz..."
    tar -xf ${APP_NAME}.tar.xz
fi

# ============================================================
# Step 3: Build binary with AppID if missing
# ============================================================
if [ ! -f "${APP_NAME}" ]; then
    echo "⚙️  Building binary with AppID: ${APP_ID}..."
    go build -ldflags="-X 'main.AppID=${APP_ID}'" -o ${APP_NAME}
fi

# ============================================================
# Step 4: Create .deb package with full icon support
# ============================================================
echo "📦 Creating .deb package..."
mkdir -p ${APP_NAME}-deb/DEBIAN
mkdir -p ${APP_NAME}-deb/usr/local/bin
mkdir -p ${APP_NAME}-deb/usr/share/applications

# Copy binary
cp ${APP_NAME} ${APP_NAME}-deb/usr/local/bin/

# Ensure icon.png is included in all standard icon sizes (tray + menu)
ICON_SIZES=(16 22 24 32 48 64 128 256)
for SIZE in "${ICON_SIZES[@]}"; do
    ICON_DIR=${APP_NAME}-deb/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps
    mkdir -p ${ICON_DIR}
    cp icon.png ${ICON_DIR}/${APP_NAME}.png
done

# Control file
cat <<EOF > ${APP_NAME}-deb/DEBIAN/control
Package: ${APP_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${AUTHOR} <${EMAIL}>
Homepage: ${REPO}
Description: ${DESCRIPTION_SHORT}
 ${DESCRIPTION_LONG}
EOF

# Desktop entry (include X-AppID for tray icon)
cat <<EOF > ${APP_NAME}-deb/usr/share/applications/${APP_NAME}.desktop
[Desktop Entry]
Name=FyClip
Exec=/usr/local/bin/${APP_NAME}
Icon=${APP_NAME}
Type=Application
Categories=Utility;
Comment=${DESCRIPTION_LONG}
StartupNotify=true
X-GNOME-UsesNotifications=true
X-AppID=${APP_ID}
EOF

# Build .deb
dpkg-deb --build ${APP_NAME}-deb
mv ${APP_NAME}-deb.deb ${APP_NAME}_${VERSION}_${ARCH}.deb
echo "✅ .deb package created: ${APP_NAME}_${VERSION}_${ARCH}.deb"

# ============================================================
# Step 5: Create AppImage (with AppID support)
# ============================================================
echo "📦 Creating AppImage..."
mkdir -p FyClip.AppDir/usr/bin
mkdir -p FyClip.AppDir/usr/share/icons/hicolor
mkdir -p FyClip.AppDir/usr/share/applications

# Copy binary and create AppRun
cp ${APP_NAME} FyClip.AppDir/
chmod +x FyClip.AppDir/${APP_NAME}
ln -sf ${APP_NAME} FyClip.AppDir/AppRun

# Desktop entry in AppDir
cat <<EOF > FyClip.AppDir/${APP_NAME}.desktop
[Desktop Entry]
Name=FyClip
Exec=${APP_NAME}
Icon=${APP_NAME}
Type=Application
Categories=Utility;
Comment=${DESCRIPTION_LONG}
StartupNotify=true
X-GNOME-UsesNotifications=true
X-AppID=${APP_ID}
EOF

# Copy icon.png to AppDir root for tray support
cp icon.png FyClip.AppDir/${APP_NAME}.png

# Resize icons for menus
for SIZE in "${ICON_SIZES[@]}"; do
    mkdir -p FyClip.AppDir/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps
    convert icon.png -resize ${SIZE}x${SIZE} FyClip.AppDir/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps/${APP_NAME}.png
done

# Download appimagetool if missing
if ! command -v appimagetool &> /dev/null; then
    echo "⬇️ Downloading appimagetool..."
    wget -q https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage -O appimagetool
    chmod +x appimagetool
    sudo mv appimagetool /usr/local/bin/
fi

# Build AppImage
appimagetool FyClip.AppDir
mv FyClip-${APPIMAGE_ARCH}.AppImage ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage
echo "✅ AppImage created: ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

echo "🎉 All done! Packages:"
echo "  - Debian package: ${APP_NAME}_${VERSION}_${ARCH}.deb"
echo "  - Universal AppImage: ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

