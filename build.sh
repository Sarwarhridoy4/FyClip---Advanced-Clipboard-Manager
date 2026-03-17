#!/bin/bash
# =====================================================================
# FyClip Build & Package Script (Debian + AppImage)
# FIXED: stable binary name (fyclip) + correct desktop Exec handling
# =====================================================================

set -euo pipefail

APP_NAME="FyClip Clipboard Manager"   # Human-readable name
BIN_NAME="fyclip"                     # Stable executable name (no spaces)
APP_ID="com.sarwar.fyclip"
PKG_NAME="fyclip"
AUTHOR="Sarwar Hossain"
EMAIL="sarwarhridoy4@gmail.com"
DIST_DIR="dist"
WORK_DIR="${DIST_DIR}/work"
FYNE_ROOT="${WORK_DIR}/fyne-root"
DEB_ROOT="${WORK_DIR}/deb-root"
APPDIR="${WORK_DIR}/FyClip.AppDir"
LOCAL_BIN_DIR="$(pwd)/${DIST_DIR}/.tools/bin"

PM=""

detect_pm() {
    if [ -n "${PM}" ]; then return; fi
    for pm in apt-get dnf pacman zypper brew; do
        if command -v "${pm}" >/dev/null 2>&1; then
            PM="${pm}"; return
        fi
    done
    PM="none"
}

run_install_cmd() {
    local install_cmd="$1"
    echo "🔧 Installing via: ${install_cmd}"
    if bash -lc "${install_cmd}"; then return 0; fi
    if command -v sudo >/dev/null 2>&1; then
        echo "🔧 Retrying with sudo..."
        sudo bash -lc "${install_cmd}"; return $?
    fi
    return 1
}

install_appimagetool_local() {
    local arch="${APPIMAGE_ARCH:-x86_64}"
    local url="https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${arch}.AppImage"
    local tmp_file
    mkdir -p "${LOCAL_BIN_DIR}"
    tmp_file="$(mktemp)"

    echo "🔧 Downloading appimagetool (${arch})..."
    if command -v curl >/dev/null 2>&1; then
        curl -LfsS "${url}" -o "${tmp_file}" || return 1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "${tmp_file}" "${url}" || return 1
    else
        if ! install_with_pm curl && ! install_with_pm wget; then
            rm -f "${tmp_file}"; return 1
        fi
        curl -LfsS "${url}" -o "${tmp_file}" || return 1
    fi

    mv "${tmp_file}" "${LOCAL_BIN_DIR}/appimagetool"
    chmod +x "${LOCAL_BIN_DIR}/appimagetool"
    export PATH="${LOCAL_BIN_DIR}:${PATH}"
}

install_with_pm() {
    local pkg="$1"
    detect_pm
    case "${PM}" in
      apt-get) run_install_cmd "apt-get update && apt-get install -y ${pkg}" ;;
      dnf) run_install_cmd "dnf install -y ${pkg}" ;;
      pacman) run_install_cmd "pacman -Sy --noconfirm ${pkg}" ;;
      zypper) run_install_cmd "zypper --non-interactive install ${pkg}" ;;
      brew) run_install_cmd "brew install ${pkg}" ;;
      *) return 1 ;;
    esac
}

install_tool() {
    local cmd="$1"
    case "${cmd}" in
      fyne)
        if command -v go >/dev/null 2>&1; then
            echo "🔧 Installing fyne CLI..."
            go install fyne.io/tools/cmd/fyne@latest && export PATH="$(go env GOPATH)/bin:${PATH}" && return 0
        fi
        install_with_pm fyne && return 0
        ;;
      go) install_with_pm golang-go || install_with_pm golang ;;
      appimagetool) install_with_pm appimagetool || install_appimagetool_local ;;
      dpkg-deb) install_with_pm dpkg ;;
      ldd) install_with_pm libc-bin || install_with_pm glibc ;;
      tar) install_with_pm tar ;;
      *) return 1 ;;
    esac
}

ensure_cmd() {
    local cmd="$1"
    if command -v "${cmd}" >/dev/null 2>&1; then return 0; fi
    echo "Missing required tool: ${cmd}"
    if install_tool "${cmd}" && command -v "${cmd}" >/dev/null 2>&1; then
        echo "✅ Installed ${cmd}"; return 0
    fi
    echo "❌ Could not auto-install '${cmd}'" >&2; exit 1
}

# ---------------------------------------------------------------------
# Version
# ---------------------------------------------------------------------
if git rev-parse --git-dir >/dev/null 2>&1; then
    DEFAULT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "1.0.0")
else
    DEFAULT_VERSION="1.0.0"
fi

if [ -n "${1:-}" ]; then VERSION="$1";
else read -r -p "Enter version [${DEFAULT_VERSION}]: " VERSION; VERSION=${VERSION:-$DEFAULT_VERSION}; fi

ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64) ARCH="amd64"; APPIMAGE_ARCH="x86_64" ;;
  aarch64) ARCH="arm64"; APPIMAGE_ARCH="aarch64" ;;
  armv7l) ARCH="armhf"; APPIMAGE_ARCH="armhf" ;;
  i386|i686) ARCH="i386"; APPIMAGE_ARCH="i686" ;;
  *) ARCH="$ARCH_RAW"; APPIMAGE_ARCH="$ARCH_RAW" ;;
esac

echo "📦 Building ${APP_NAME} ${VERSION} (${ARCH})"

ensure_cmd go
ensure_cmd fyne
ensure_cmd tar
ensure_cmd dpkg-deb
ensure_cmd appimagetool
ensure_cmd ldd

mkdir -p "${DIST_DIR}" "${WORK_DIR}"
rm -rf "${FYNE_ROOT}" "${DEB_ROOT}" "${APPDIR}"
rm -f "${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb" "${DIST_DIR}/${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

# ---------------------------------------------------------------------
# Fyne package
# ---------------------------------------------------------------------
rm -f "${APP_NAME}.tar.xz"
fyne package --os linux --release --name "${APP_NAME}" --icon icon.png

tar -xJf "${APP_NAME}.tar.xz" -C "${WORK_DIR}"

# Detect prefix
if [ -d "${WORK_DIR}/usr/local" ]; then PREFIX_REL="usr/local";
else PREFIX_REL="usr"; fi

USR_NORMALIZED="${WORK_DIR}/usr-normalized"
mkdir -p "${USR_NORMALIZED}"
cp -a "${WORK_DIR}/${PREFIX_REL}/." "${USR_NORMALIZED}/"

BIN_DIR="${USR_NORMALIZED}/bin"
APPS_DIR="${USR_NORMALIZED}/share/applications"
PIXMAPS_DIR="${USR_NORMALIZED}/share/pixmaps"

# Normalize binary name → fyclip
FOUND_BIN=$(find "${BIN_DIR}" -maxdepth 1 -type f -executable | head -n 1)
if [ -z "${FOUND_BIN}" ]; then echo "Binary not found"; exit 1; fi
mv "${FOUND_BIN}" "${BIN_DIR}/${BIN_NAME}"
BIN_PATH="${BIN_DIR}/${BIN_NAME}"

# Desktop + icon
DESKTOP_PATH=$(find "${APPS_DIR}" -name '*.desktop' | head -n 1)
ICON_PATH=$(find "${PIXMAPS_DIR}" -type f | head -n 1)

if [ -z "${DESKTOP_PATH}" ]; then echo "Desktop file not found"; exit 1; fi
if [ -z "${ICON_PATH}" ]; then echo "Icon file not found"; exit 1; fi

# GNOME Wayland groups windows to desktop entries by desktop file id (filename without .desktop).
# Normalize both the desktop filename and icon filename to match our app id.
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

sed -i -E "s|^Exec=.*|Exec=${BIN_NAME}|" "${DESKTOP_PATH}"
sed -i -E "s|^Icon=.*|Icon=${APP_ID}|" "${DESKTOP_PATH}"
sed -i -E "s|^Name=.*|Name=${APP_NAME}|" "${DESKTOP_PATH}"

grep -q '^Categories=' "${DESKTOP_PATH}" || echo "Categories=Utility;" >> "${DESKTOP_PATH}"
grep -q '^NoDisplay=' "${DESKTOP_PATH}" || echo "NoDisplay=false" >> "${DESKTOP_PATH}"
grep -q '^Keywords=' "${DESKTOP_PATH}" || echo "Keywords=clipboard;copy;paste;history;" >> "${DESKTOP_PATH}"
grep -q '^StartupWMClass=' "${DESKTOP_PATH}" || echo "StartupWMClass=${APP_ID}" >> "${DESKTOP_PATH}"

# Install icon into hicolor so `gtk-update-icon-cache` picks it up on Debian-based distros.
HICOLOR_APPS_DIR="${USR_NORMALIZED}/share/icons/hicolor/256x256/apps"
mkdir -p "${HICOLOR_APPS_DIR}"
cp -f "${ICON_PATH}" "${HICOLOR_APPS_DIR}/${APP_ID}.${ICON_EXT}"

# ---------------------------------------------------------------------
# Debian
# ---------------------------------------------------------------------
mkdir -p "${DEB_ROOT}/DEBIAN" "${DEB_ROOT}/usr"
cp -a "${USR_NORMALIZED}/." "${DEB_ROOT}/usr/"

cat > "${DEB_ROOT}/DEBIAN/control" <<EOF
Package: ${PKG_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${AUTHOR} <${EMAIL}>
Depends: libgl1, libx11-6, libxcursor1, libxrandr2, libxinerama1, libxi6, libgtk-3-0, xclip, xsel, wl-clipboard
Description: FyClip - Advanced Clipboard Manager
EOF

cat > "${DEB_ROOT}/DEBIAN/postinst" <<'EOF'
#!/bin/sh
set -e
update-desktop-database -q || true
gtk-update-icon-cache -q /usr/share/icons/hicolor || true
EOF
chmod 755 "${DEB_ROOT}/DEBIAN/postinst"

dpkg-deb --build "${DEB_ROOT}" "${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb"

# ---------------------------------------------------------------------
# AppImage
# ---------------------------------------------------------------------
mkdir -p "${APPDIR}/usr"
cp -a "${USR_NORMALIZED}/." "${APPDIR}/usr/"

cat > "${APPDIR}/AppRun" <<EOF
#!/bin/sh
HERE="\$(dirname "\$(readlink -f "\$0")")"
exec "\$HERE/usr/bin/${BIN_NAME}" "\$@"
EOF
chmod +x "${APPDIR}/AppRun"

cp "${DESKTOP_PATH}" "${APPDIR}/${APP_ID}.desktop"
cp "${ICON_PATH}" "${APPDIR}/${APP_ID}.${ICON_EXT}"

# AppImage launchers expect Exec=AppRun.
sed -i -E "s|^Exec=.*|Exec=AppRun|" "${APPDIR}/${APP_ID}.desktop"

mkdir -p "${APPDIR}/usr/lib"
ldd "${BIN_PATH}" | awk '/=> \\// {print $3}' | xargs -r -I '{}' cp '{}' "${APPDIR}/usr/lib/" || true

APPIMAGE_EXTRACT_AND_RUN=1 appimagetool "${APPDIR}" "${DIST_DIR}/${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"

echo "🎉 Build complete"
echo "Deb: ${DIST_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb"
echo "AppImage: ${DIST_DIR}/${PKG_NAME}_${VERSION}_${APPIMAGE_ARCH}.AppImage"
