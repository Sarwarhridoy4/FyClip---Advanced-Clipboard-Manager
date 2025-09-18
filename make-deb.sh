#!/bin/bash
# =====================================================================
# FyClip Build & Package Script (Automated for Debian + AppImage)
#
# Features:
#   ✅ Tray icon support (icons in hicolor paths)
#   ✅ Multi-size screenshots (software centers)
#   ✅ Version prompt (defaults to latest Git tag or 1.0.0)
#   ✅ Git commit hash embedded in binary
#   ✅ AppID support
#   ✅ Builds .deb and AppImage packages
#   ✅ postinst script: refreshes caches after install
#   ✅ prerm script: refreshes caches after uninstall
#
# Usage:
#   ./build-all.sh [version]
#
# Author : Sarwar Hossain
# Email  : sarwarhridoy4@gmail.com
# Repo   : https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager
# =====================================================================

APP_NAME="fyclip"
APP_ID="com.sarwar.fyclip"   # Unique AppID used by Fyne

# --- Determine version from argument, Git tag, or prompt (fallback 1.0.0) ---
if git rev-parse --git-dir > /dev/null 2>&1; then
    DEFAULT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "1.0.0")
    GIT_HASH=$(git rev-parse --short HEAD)
else
    DEFAULT_VERSION="1.0.0"
    GIT_HASH="unknown"
fi

if [ -n "$1" ]; then
    VERSION="$1"
else
    read -p "Enter version [default: ${DEFAULT_VERSION}]: " VERSION
    VERSION=${VERSION:-"$DEFAULT_VERSION"}
fi

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
DESCRIPTION_SHORT="FyClip - Advanced Clipboard Manager"
DESCRIPTION_LONG="A powerful, cross-platform clipboard manager built with Go and Fyne that automatically tracks your clipboard history, provides instant search, and persists data between sessions. Now with image support and pinning for favorite items."

echo "📦 Building ${APP_NAME} version ${VERSION} (commit ${GIT_HASH}) for ${ARCH}"

# =====================================================================
# Step 0: Install required build tools
# =====================================================================
echo "🔧 Installing required tools (ImageMagick, wget, dpkg-deb, Go)..."
sudo apt-get update
sudo apt-get install -y imagemagick wget dpkg-dev golang-go
sudo go mod init ${APP_NAME}
sudo go mod tidy

# =====================================================================
# Step 1: Clean previous builds
# =====================================================================
rm -rf ${APP_NAME}-deb ${APP_NAME}_${VERSION}_${ARCH}.deb FyClip.AppDir ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage

# =====================================================================
# Step 2: Build binary with metadata
# =====================================================================
echo "⚙️  Building Go binary with AppID + Git commit..."
go build -ldflags="-X 'main.AppID=${APP_ID}' -X 'main.GitCommit=${GIT_HASH}'" -o ${APP_NAME}

# =====================================================================
# Step 3: Create Debian package
# =====================================================================
echo "📦 Creating .deb package..."

# --- Directory structure ---
mkdir -p ${APP_NAME}-deb/DEBIAN
mkdir -p ${APP_NAME}-deb/usr/bin
mkdir -p ${APP_NAME}-deb/usr/share/applications
mkdir -p ${APP_NAME}-deb/usr/share/icons/hicolor
mkdir -p ${APP_NAME}-deb/usr/share/${APP_NAME}/screenshots

# --- Copy binary ---
cp ${APP_NAME} ${APP_NAME}-deb/usr/bin/

# --- Screenshots for Software Center ---
if [ -f "screenshot.png" ]; then
    SCREENSHOT_SIZES=(320 640 1280)
    for SIZE in "${SCREENSHOT_SIZES[@]}"; do
        convert screenshot.png -resize ${SIZE}x ${APP_NAME}-deb/usr/share/${APP_NAME}/screenshots/screenshot_${SIZE}.png
    done
fi

# --- Icons (system tray + menus) ---
ICON_SIZES=(16 22 24 32 48 64 128 256 512)
for SIZE in "${ICON_SIZES[@]}"; do
    ICON_DIR=${APP_NAME}-deb/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps
    mkdir -p ${ICON_DIR}
    convert icon.png -resize ${SIZE}x${SIZE} ${ICON_DIR}/${APP_NAME}.png
done

# --- Control file ---
cat <<EOF > ${APP_NAME}-deb/DEBIAN/control
Package: ${APP_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${AUTHOR} <${EMAIL}>
Homepage: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager
Description: ${DESCRIPTION_SHORT}
 ${DESCRIPTION_LONG}
EOF

# --- postinst maintainer script ---
cat <<'EOF' > ${APP_NAME}-deb/DEBIAN/postinst
#!/bin/bash
set -e
echo "🔄 Updating icon and desktop caches (post-install)..."
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database -q
fi
if command -v gtk-update-icon-cache &> /dev/null; then
    gtk-update-icon-cache -q /usr/share/icons/hicolor
fi
exit 0
EOF
chmod 755 ${APP_NAME}-deb/DEBIAN/postinst

# --- prerm maintainer script ---
cat <<'EOF' > ${APP_NAME}-deb/DEBIAN/prerm
#!/bin/bash
set -e
echo "🗑️ Cleaning up icon and desktop caches (pre-remove)..."
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database -q
fi
if command -v gtk-update-icon-cache &> /dev/null; then
    gtk-update-icon-cache -q /usr/share/icons/hicolor
fi
exit 0
EOF
chmod 755 ${APP_NAME}-deb/DEBIAN/prerm

# --- Desktop entry ---
cat <<EOF > ${APP_NAME}-deb/usr/share/applications/${APP_NAME}.desktop
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
X-AppInstall-Screenshot=/usr/share/${APP_NAME}/screenshots/screenshot_640.png
EOF

# --- Build Debian package ---
dpkg-deb --build ${APP_NAME}-deb
mv ${APP_NAME}-deb.deb ${APP_NAME}_${VERSION}_${ARCH}.deb
echo "✅ .deb package created: ${APP_NAME}_${VERSION}_${ARCH}.deb"

# =====================================================================
# Step 4: Create AppImage
# =====================================================================
echo "📦 Creating AppImage..."

mkdir -p FyClip.AppDir/usr/bin
mkdir -p FyClip.AppDir/usr/share/icons/hicolor
mkdir -p FyClip.AppDir/usr/share/applications
mkdir -p FyClip.AppDir/usr/share/${APP_NAME}/screenshots

# --- Copy binary + AppRun ---
cp ${APP_NAME} FyClip.AppDir/${APP_NAME}
chmod +x FyClip.AppDir/${APP_NAME}
ln -sf ${APP_NAME} FyClip.AppDir/AppRun

# --- Screenshots ---
if [ -f "screenshot.png" ]; then
    for SIZE in "${SCREENSHOT_SIZES[@]}"; do
        convert screenshot.png -resize ${SIZE}x FyClip.AppDir/usr/share/${APP_NAME}/screenshots/screenshot_${SIZE}.png
    done
fi

# --- Desktop entry ---
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
X-AppInstall-Screenshot=/usr/share/${APP_NAME}/screenshots/screenshot_640.png
EOF

# --- Icons (including tray support) ---
cp icon.png FyClip.AppDir/${APP_NAME}.png
for SIZE in "${ICON_SIZES[@]}"; do
    mkdir -p FyClip.AppDir/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps
    convert icon.png -resize ${SIZE}x${SIZE} FyClip.AppDir/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps/${APP_NAME}.png
done

# --- AppImageTool ---
if ! command -v appimagetool &> /dev/null; then
    echo "⬇️ Downloading appimagetool..."
    wget -q https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${APPIMAGE_ARCH}.AppImage -O appimagetool
    chmod +x appimagetool
    sudo mv appimagetool /usr/local/bin/
fi

# --- Build AppImage ---
appimagetool FyClip.AppDir
mv FyClip-${APPIMAGE_ARCH}.AppImage ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage
echo "✅ AppImage created: ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

# =====================================================================
# Done
# =====================================================================
echo "🎉 Build complete! Packages created:"
echo "  • Debian package : ${APP_NAME}_${VERSION}_${ARCH}.deb"
echo "  • AppImage       : ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"
