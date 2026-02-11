#!/bin/bash
# =====================================================================
# FyClip Build & Package Script (Debian + AppImage)
# Uses Fyne's official Linux package output as the base payload.
# =====================================================================

set -euo pipefail

APP_NAME="fyclip"
APP_ID="com.sarwar.fyclip"
AUTHOR="Sarwar Hossain"
EMAIL="sarwarhridoy4@gmail.com"
DIST_DIR="dist"
WORK_DIR="${DIST_DIR}/work"
FYNE_ROOT="${WORK_DIR}/fyne-root"
DEB_ROOT="${WORK_DIR}/deb-root"
APPDIR="${WORK_DIR}/FyClip.AppDir"

require_cmd() {
    local cmd="$1"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Missing required tool: ${cmd}" >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------
# Version detection
# ---------------------------------------------------------------------
if git rev-parse --git-dir >/dev/null 2>&1; then
    DEFAULT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "1.0.0")
else
    DEFAULT_VERSION="1.0.0"
fi

if [ -n "${1:-}" ]; then
    VERSION="$1"
else
    read -r -p "Enter version [default: ${DEFAULT_VERSION}]: " VERSION
    VERSION=${VERSION:-"$DEFAULT_VERSION"}
fi

ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64) ARCH="amd64"; APPIMAGE_ARCH="x86_64" ;;
  aarch64) ARCH="arm64"; APPIMAGE_ARCH="aarch64" ;;
  armv7l) ARCH="armhf"; APPIMAGE_ARCH="armhf" ;;
  i386|i686) ARCH="i386"; APPIMAGE_ARCH="i686" ;;
  *) ARCH="$ARCH_RAW"; APPIMAGE_ARCH="$ARCH_RAW" ;;
esac

echo "📦 Building ${APP_NAME} ${VERSION} for ${ARCH}"

# ---------------------------------------------------------------------
# Requirements
# ---------------------------------------------------------------------
require_cmd go
require_cmd fyne
require_cmd tar
require_cmd dpkg-deb
require_cmd appimagetool
require_cmd ldd

# ---------------------------------------------------------------------
# Build official Fyne Linux package payload
# ---------------------------------------------------------------------
mkdir -p "${DIST_DIR}" "${WORK_DIR}"
rm -rf "${FYNE_ROOT}" "${DEB_ROOT}" "${APPDIR}"
rm -f "${DIST_DIR}/${APP_NAME}_${VERSION}_${ARCH}.deb" "${DIST_DIR}/${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

echo "⚙️ Running official Fyne Linux package step..."
rm -f "${APP_NAME}.tar.xz"
fyne package --os linux --release --name "${APP_NAME}" --icon icon.png

if [ ! -f "${APP_NAME}.tar.xz" ]; then
    echo "Fyne packaging failed: ${APP_NAME}.tar.xz not found" >&2
    exit 1
fi

mkdir -p "${FYNE_ROOT}"
tar -xJf "${APP_NAME}.tar.xz" -C "${FYNE_ROOT}"

if [ -d "${FYNE_ROOT}/usr/local" ]; then
    PREFIX_REL="usr/local"
elif [ -d "${FYNE_ROOT}/usr" ]; then
    PREFIX_REL="usr"
else
    echo "Unexpected Fyne payload layout: missing usr or usr/local directories" >&2
    exit 1
fi

BIN_PATH="${FYNE_ROOT}/${PREFIX_REL}/bin/${APP_NAME}"
DESKTOP_PATH="${FYNE_ROOT}/${PREFIX_REL}/share/applications/${APP_ID}.desktop"
ICON_PATH="${FYNE_ROOT}/${PREFIX_REL}/share/pixmaps/${APP_ID}.png"

if [ ! -f "${BIN_PATH}" ] || [ ! -f "${DESKTOP_PATH}" ] || [ ! -f "${ICON_PATH}" ]; then
    echo "Fyne payload missing expected binary/desktop/icon assets" >&2
    exit 1
fi

# ---------------------------------------------------------------------
# Debian package from official payload
# ---------------------------------------------------------------------
echo "📦 Creating Debian package..."
mkdir -p "${DEB_ROOT}/DEBIAN"
cp -a "${FYNE_ROOT}/usr" "${DEB_ROOT}/"

cat <<EOF > "${DEB_ROOT}/DEBIAN/control"
Package: ${APP_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${AUTHOR} <${EMAIL}>
Depends: libgl1, libx11-6, libxcursor1, libxrandr2, libxinerama1, libxi6, libgtk-3-0, xclip, xsel, wl-clipboard
Description: FyClip - Advanced Clipboard Manager
EOF

cat <<'EOF' > "${DEB_ROOT}/DEBIAN/postinst"
#!/bin/sh
set -e
command -v update-desktop-database >/dev/null && update-desktop-database -q
command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q /usr/share/icons/hicolor || true
EOF
chmod 755 "${DEB_ROOT}/DEBIAN/postinst"

dpkg-deb --build "${DEB_ROOT}" "${DIST_DIR}/${APP_NAME}_${VERSION}_${ARCH}.deb"

# ---------------------------------------------------------------------
# AppImage from official payload
# ---------------------------------------------------------------------
echo "📦 Creating AppImage..."
mkdir -p "${APPDIR}"
cp -a "${FYNE_ROOT}/usr" "${APPDIR}/"

cat > "${APPDIR}/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
if [ -x "$HERE/usr/local/bin/fyclip" ]; then
  exec "$HERE/usr/local/bin/fyclip" "$@"
fi
exec "$HERE/usr/bin/fyclip" "$@"
EOF
chmod +x "${APPDIR}/AppRun"

cp "${DESKTOP_PATH}" "${APPDIR}/${APP_ID}.desktop"
cp "${ICON_PATH}" "${APPDIR}/${APP_ID}.png"

mkdir -p "${APPDIR}/usr/lib"
ldd "${BIN_PATH}" | awk '/=> \// {print $3}' | xargs -r -I '{}' cp -v '{}' "${APPDIR}/usr/lib/" || true

# Work in environments without FUSE, e.g. CI/sandboxes.
APPIMAGE_EXTRACT_AND_RUN=1 appimagetool "${APPDIR}" "${DIST_DIR}/${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

echo "🎉 Build complete"
echo " • Debian  : ${DIST_DIR}/${APP_NAME}_${VERSION}_${ARCH}.deb"
echo " • AppImage: ${DIST_DIR}/${APP_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"
