#!/bin/bash
# =====================================================================
# FyClip Build & Package Script (Debian + AppImage)
# Fully automated: installs missing dependencies, fixes icons, cross-distro
# =====================================================================

APP_NAME="fyclip"
APP_ID="com.sarwar.fyclip"
ICON_NAME="fyclip"
AUTHOR="Sarwar Hossain"
EMAIL="sarwarhridoy4@gmail.com"

# --- Determine version from argument, Git tag, or fallback ---
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

ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64) ARCH="amd64"; APPIMAGE_ARCH="x86_64" ;;
  aarch64) ARCH="arm64"; APPIMAGE_ARCH="arm64" ;;
  armv7l) ARCH="armhf"; APPIMAGE_ARCH="armhf" ;;
  i386) ARCH="i386"; APPIMAGE_ARCH="i386" ;;
  *) ARCH="$ARCH_RAW"; APPIMAGE_ARCH="$ARCH_RAW" ;;
esac

echo "📦 Building ${APP_NAME} version ${VERSION} (commit ${GIT_HASH}) for ${ARCH}"

# =====================================================================
# Step 0: Function to auto-install missing dependencies
# =====================================================================
install_if_missing() {
    local cmd="$1"
    local pkg="$2"
    if ! command -v "$cmd" &> /dev/null; then
        echo "⬇️ $cmd not found. Installing $pkg..."
        sudo apt-get update
        sudo apt-get install -y "$pkg"
    else
        echo "✅ $cmd is already installed"
    fi
}

# Install build tools and runtime dependencies
install_if_missing go golang-go
install_if_missing magick imagemagick
install_if_missing dpkg-deb dpkg-dev
install_if_missing xclip xclip
install_if_missing xsel xsel
install_if_missing wl-copy wl-clipboard
install_if_missing git git
install_if_missing unzip unzip
install_if_missing curl curl

# AppImage tool (download if missing)
if ! command -v appimagetool &> /dev/null; then
    echo "⬇️ Downloading appimagetool..."
    wget -q https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${APPIMAGE_ARCH}.AppImage -O appimagetool
    chmod +x appimagetool
    sudo mv appimagetool /usr/local/bin/
fi

# =====================================================================
# Step 0b: Detect ImageMagick command
# =====================================================================
if command -v magick &> /dev/null; then
    IMAGEMAGICK_CMD="magick convert"
elif command -v convert &> /dev/null; then
    IMAGEMAGICK_CMD="convert"
else
    echo "❌ ImageMagick not found. Please install it."
    exit 1
fi

# =====================================================================
# Step 1: Initialize Go module
# =====================================================================
if [ ! -f "go.mod" ]; then
    go mod init ${APP_NAME}
fi
go mod tidy

# =====================================================================
# Step 2: Clean previous builds
# =====================================================================
rm -rf ${APP_NAME}-deb ${APP_NAME}_${VERSION}_${ARCH}.deb FyClip.AppDir ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage

# =====================================================================
# Step 3: Build Go binary with metadata
# =====================================================================
echo "⚙️  Building Go binary with AppID + Git commit..."
go build -ldflags="-X 'main.AppID=${APP_ID}' -X 'main.GitCommit=${GIT_HASH}'" -o ${APP_NAME}

# =====================================================================
# Step 4: Create Debian package
# =====================================================================
echo "📦 Creating .deb package..."
mkdir -p ${APP_NAME}-deb/DEBIAN
mkdir -p ${APP_NAME}-deb/usr/bin
mkdir -p ${APP_NAME}-deb/usr/share/applications
mkdir -p ${APP_NAME}-deb/usr/share/icons/hicolor
mkdir -p ${APP_NAME}-deb/usr/share/pixmaps
mkdir -p ${APP_NAME}-deb/usr/share/${APP_NAME}/screenshots

cp ${APP_NAME} ${APP_NAME}-deb/usr/bin/

# Screenshots
if [ -f "screenshot.png" ]; then
    SCREENSHOT_SIZES=(320 640 1280)
    for SIZE in "${SCREENSHOT_SIZES[@]}"; do
        $IMAGEMAGICK_CMD screenshot.png -resize ${SIZE}x ${APP_NAME}-deb/usr/share/${APP_NAME}/screenshots/screenshot_${SIZE}.png
    done
fi

# Icons
ICON_SIZES=(16 22 24 32 48 64 128 256 512)
for SIZE in "${ICON_SIZES[@]}"; do
    ICON_DIR=${APP_NAME}-deb/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps
    mkdir -p ${ICON_DIR}
    $IMAGEMAGICK_CMD icon.png -resize ${SIZE}x${SIZE} ${ICON_DIR}/${ICON_NAME}.png
done
cp icon.png ${APP_NAME}-deb/usr/share/pixmaps/${ICON_NAME}.png

# Control file
cat <<EOF > ${APP_NAME}-deb/DEBIAN/control
Package: ${APP_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${AUTHOR} <${EMAIL}>
Homepage: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager
Depends: libgl1, libx11-6, libxcursor1, libxrandr2, libxinerama1, libxi6, libxxf86vm1, libgtk-3-0, libappindicator3-1, libwebkit2gtk-4.1-0, xclip, xsel, wl-clipboard
Description: FyClip - Advanced Clipboard Manager
 A powerful, cross-platform clipboard manager built with Go and Fyne that tracks clipboard history, supports images, and persists data.
EOF

# postinst
cat <<'EOF' > ${APP_NAME}-deb/DEBIAN/postinst
#!/bin/bash
set -e
if command -v update-desktop-database &> /dev/null; then update-desktop-database -q; fi
if command -v gtk-update-icon-cache &> /dev/null; then gtk-update-icon-cache -q /usr/share/icons/hicolor; fi
exit 0
EOF
chmod 755 ${APP_NAME}-deb/DEBIAN/postinst

# prerm
cat <<'EOF' > ${APP_NAME}-deb/DEBIAN/prerm
#!/bin/bash
set -e
if command -v update-desktop-database &> /dev/null; then update-desktop-database -q; fi
if command -v gtk-update-icon-cache &> /dev/null; then gtk-update-icon-cache -q /usr/share/icons/hicolor; fi
exit 0
EOF
chmod 755 ${APP_NAME}-deb/DEBIAN/prerm

# Desktop entry
cat <<EOF > ${APP_NAME}-deb/usr/share/applications/${APP_NAME}.desktop
[Desktop Entry]
Name=FyClip
Exec=${APP_NAME}
Icon=${ICON_NAME}
Type=Application
Categories=Utility;
Comment=A cross-platform clipboard manager with image support
StartupNotify=true
X-GNOME-UsesNotifications=true
X-AppID=${APP_ID}
X-AppInstall-Screenshot=/usr/share/${APP_NAME}/screenshots/screenshot_640.png
EOF

dpkg-deb --build ${APP_NAME}-deb
mv ${APP_NAME}-deb.deb ${APP_NAME}_${VERSION}_${ARCH}.deb
echo "✅ .deb package created: ${APP_NAME}_${VERSION}_${ARCH}.deb"

# =====================================================================
# Step 5: Create AppImage
# =====================================================================
echo "📦 Creating AppImage..."
mkdir -p FyClip.AppDir/usr/bin
mkdir -p FyClip.AppDir/usr/share/icons/hicolor
mkdir -p FyClip.AppDir/usr/share/applications
mkdir -p FyClip.AppDir/usr/share/pixmaps
mkdir -p FyClip.AppDir/usr/share/${APP_NAME}/screenshots

cp ${APP_NAME} FyClip.AppDir/usr/bin/
chmod +x FyClip.AppDir/usr/bin/${APP_NAME}
ln -sf usr/bin/${APP_NAME} FyClip.AppDir/AppRun

# Screenshots
if [ -f "screenshot.png" ]; then
    for SIZE in "${SCREENSHOT_SIZES[@]}"; do
        $IMAGEMAGICK_CMD screenshot.png -resize ${SIZE}x FyClip.AppDir/usr/share/${APP_NAME}/screenshots/screenshot_${SIZE}.png
    done
fi

# Icons
for SIZE in "${ICON_SIZES[@]}"; do
    ICON_DIR=FyClip.AppDir/usr/share/icons/hicolor/${SIZE}x${SIZE}/apps
    mkdir -p ${ICON_DIR}
    $IMAGEMAGICK_CMD icon.png -resize ${SIZE}x${SIZE} ${ICON_DIR}/${ICON_NAME}.png
done

# Copy main icon to AppDir root for AppImage
cp icon.png FyClip.AppDir/${ICON_NAME}.png
cp icon.png FyClip.AppDir/usr/share/pixmaps/${ICON_NAME}.png

# Desktop entry
cat <<EOF > FyClip.AppDir/${APP_NAME}.desktop
[Desktop Entry]
Name=FyClip
Exec=${APP_NAME}
Icon=${ICON_NAME}
Type=Application
Categories=Utility;
Comment=A cross-platform clipboard manager with image support
StartupNotify=true
X-GNOME-UsesNotifications=true
X-AppID=${APP_ID}
X-AppInstall-Screenshot=/usr/share/${APP_NAME}/screenshots/screenshot_640.png
EOF

# Bundle required shared libraries
mkdir -p FyClip.AppDir/usr/lib
ldd ${APP_NAME} | grep "=> /" | awk '{print $3}' | xargs -I '{}' cp -v '{}' FyClip.AppDir/usr/lib/ || true

# Build AppImage
appimagetool FyClip.AppDir
mv FyClip-${APPIMAGE_ARCH}.AppImage ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage
echo "✅ AppImage created: ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

# =====================================================================
echo "🎉 Build complete!"
echo "  • Debian package : ${APP_NAME}_${VERSION}_${ARCH}.deb"
echo "  • AppImage       : ${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"