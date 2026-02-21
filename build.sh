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
LOCAL_BIN_DIR="$(pwd)/${DIST_DIR}/.tools/bin"

PM=""

detect_pm() {
    if [ -n "${PM}" ]; then
        return
    fi
    for pm in apt-get dnf pacman zypper brew; do
        if command -v "${pm}" >/dev/null 2>&1; then
            PM="${pm}"
            return
        fi
    done
    PM="none"
}

run_install_cmd() {
    local install_cmd="$1"
    echo "🔧 Installing via: ${install_cmd}"
    if bash -lc "${install_cmd}"; then
        return 0
    fi
    if command -v sudo >/dev/null 2>&1; then
        echo "🔧 Retrying with sudo..."
        sudo bash -lc "${install_cmd}"
        return $?
    fi
    return 1
}

install_appimagetool_local() {
    local arch="${APPIMAGE_ARCH:-x86_64}"
    local url="https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${arch}.AppImage"
    local tmp_file
    mkdir -p "${LOCAL_BIN_DIR}"
    tmp_file="$(mktemp)"

    echo "🔧 Downloading appimagetool (${arch}) from AppImageKit..."
    if command -v curl >/dev/null 2>&1; then
        curl -LfsS "${url}" -o "${tmp_file}" || return 1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "${tmp_file}" "${url}" || return 1
    else
        if ! install_with_pm curl && ! install_with_pm wget; then
            rm -f "${tmp_file}"
            return 1
        fi
        if command -v curl >/dev/null 2>&1; then
            curl -LfsS "${url}" -o "${tmp_file}" || return 1
        else
            wget -qO "${tmp_file}" "${url}" || return 1
        fi
    fi

    mv "${tmp_file}" "${LOCAL_BIN_DIR}/appimagetool"
    chmod +x "${LOCAL_BIN_DIR}/appimagetool"
    export PATH="${LOCAL_BIN_DIR}:${PATH}"
    return 0
}

install_with_pm() {
    local pkg="$1"
    detect_pm
    case "${PM}" in
      apt-get)
        run_install_cmd "apt-get update && apt-get install -y ${pkg}"
        ;;
      dnf)
        run_install_cmd "dnf install -y ${pkg}"
        ;;
      pacman)
        run_install_cmd "pacman -Sy --noconfirm ${pkg}"
        ;;
      zypper)
        run_install_cmd "zypper --non-interactive install ${pkg}"
        ;;
      brew)
        run_install_cmd "brew install ${pkg}"
        ;;
      *)
        return 1
        ;;
    esac
}

install_tool() {
    local cmd="$1"
    case "${cmd}" in
      fyne)
        if command -v go >/dev/null 2>&1; then
            echo "🔧 Installing fyne CLI with go install..."
            if go install fyne.io/tools/cmd/fyne@latest; then
                export PATH="$(go env GOPATH)/bin:${PATH}"
                return 0
            fi
        fi
        if install_with_pm fyne; then
            return 0
        fi
        ;;
      go)
        if install_with_pm golang-go || install_with_pm golang; then
            return 0
        fi
        ;;
      appimagetool)
        if install_with_pm appimagetool || install_appimagetool_local; then
            return 0
        fi
        ;;
      dpkg-deb)
        if install_with_pm dpkg; then
            return 0
        fi
        ;;
      ldd)
        if install_with_pm libc-bin || install_with_pm glibc; then
            return 0
        fi
        ;;
      tar)
        if install_with_pm tar; then
            return 0
        fi
        ;;
      *)
        return 1
        ;;
    esac
    return 1
}

ensure_cmd() {
    local cmd="$1"
    if command -v "${cmd}" >/dev/null 2>&1; then
        return 0
    fi
    echo "Missing required tool: ${cmd}"
    if install_tool "${cmd}" && command -v "${cmd}" >/dev/null 2>&1; then
        echo "✅ Installed ${cmd}"
        return 0
    fi
    echo "❌ Could not auto-install '${cmd}'. Please install it manually and rerun." >&2
    exit 1
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
ensure_cmd go
ensure_cmd fyne
ensure_cmd tar
ensure_cmd dpkg-deb
ensure_cmd appimagetool
ensure_cmd ldd

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

BIN_DIR="${FYNE_ROOT}/${PREFIX_REL}/bin"
APPS_DIR="${FYNE_ROOT}/${PREFIX_REL}/share/applications"
PIXMAPS_DIR="${FYNE_ROOT}/${PREFIX_REL}/share/pixmaps"

BIN_PATH="${BIN_DIR}/${APP_NAME}"
if [ ! -x "${BIN_PATH}" ]; then
    BIN_PATH="$(find "${BIN_DIR}" -maxdepth 1 -type f -executable | head -n 1 || true)"
fi

# Force a stable command name in packaged artifacts, even if Fyne emits
# a different executable basename (e.g. repo/module name).
if [ -n "${BIN_PATH}" ] && [ -f "${BIN_PATH}" ] && [ "$(basename "${BIN_PATH}")" != "${APP_NAME}" ]; then
    rm -f "${BIN_DIR}/${APP_NAME}"
    mv "${BIN_PATH}" "${BIN_DIR}/${APP_NAME}"
    BIN_PATH="${BIN_DIR}/${APP_NAME}"
fi

DESKTOP_PATH="${APPS_DIR}/${APP_ID}.desktop"
if [ ! -f "${DESKTOP_PATH}" ]; then
    DESKTOP_PATH="$(find "${APPS_DIR}" -maxdepth 1 -type f -name '*.desktop' | head -n 1 || true)"
fi

ICON_PATH="${PIXMAPS_DIR}/${APP_ID}.png"
if [ ! -f "${ICON_PATH}" ]; then
    ICON_PATH="$(find "${PIXMAPS_DIR}" -maxdepth 1 -type f \( -name '*.png' -o -name '*.svg' -o -name '*.xpm' \) | head -n 1 || true)"
fi

if [ -z "${BIN_PATH}" ] || [ ! -f "${BIN_PATH}" ] || [ -z "${DESKTOP_PATH}" ] || [ ! -f "${DESKTOP_PATH}" ] || [ -z "${ICON_PATH}" ] || [ ! -f "${ICON_PATH}" ]; then
    echo "Fyne payload missing expected binary/desktop/icon assets" >&2
    echo "Detected values:" >&2
    echo "  BIN_PATH=${BIN_PATH:-<none>}" >&2
    echo "  DESKTOP_PATH=${DESKTOP_PATH:-<none>}" >&2
    echo "  ICON_PATH=${ICON_PATH:-<none>}" >&2
    exit 1
fi

# ---------------------------------------------------------------------
# Normalize payload to /usr and ensure desktop entry is menu-visible
# ---------------------------------------------------------------------
USR_NORMALIZED="${WORK_DIR}/usr-normalized"
rm -rf "${USR_NORMALIZED}"
mkdir -p "${USR_NORMALIZED}"

if [ "${PREFIX_REL}" = "usr/local" ]; then
    cp -a "${FYNE_ROOT}/usr/local/." "${USR_NORMALIZED}/"
else
    cp -a "${FYNE_ROOT}/usr/." "${USR_NORMALIZED}/"
fi

BIN_DIR="${USR_NORMALIZED}/bin"
APPS_DIR="${USR_NORMALIZED}/share/applications"
PIXMAPS_DIR="${USR_NORMALIZED}/share/pixmaps"

BIN_PATH="${BIN_DIR}/${APP_NAME}"
if [ ! -x "${BIN_PATH}" ]; then
    BIN_PATH="$(find "${BIN_DIR}" -maxdepth 1 -type f -executable | head -n 1 || true)"
fi
if [ -n "${BIN_PATH}" ] && [ -f "${BIN_PATH}" ] && [ "$(basename "${BIN_PATH}")" != "${APP_NAME}" ]; then
    rm -f "${BIN_DIR}/${APP_NAME}"
    mv "${BIN_PATH}" "${BIN_DIR}/${APP_NAME}"
    BIN_PATH="${BIN_DIR}/${APP_NAME}"
fi

DESKTOP_PATH="${APPS_DIR}/${APP_ID}.desktop"
if [ ! -f "${DESKTOP_PATH}" ]; then
    DESKTOP_PATH="$(find "${APPS_DIR}" -maxdepth 1 -type f -name '*.desktop' | head -n 1 || true)"
fi

ICON_PATH="${PIXMAPS_DIR}/${APP_ID}.png"
if [ ! -f "${ICON_PATH}" ]; then
    ICON_PATH="$(find "${PIXMAPS_DIR}" -maxdepth 1 -type f \( -name '*.png' -o -name '*.svg' -o -name '*.xpm' \) | head -n 1 || true)"
fi

if [ -z "${BIN_PATH}" ] || [ ! -f "${BIN_PATH}" ] || [ -z "${DESKTOP_PATH}" ] || [ ! -f "${DESKTOP_PATH}" ] || [ -z "${ICON_PATH}" ] || [ ! -f "${ICON_PATH}" ]; then
    echo "Normalized payload missing expected binary/desktop/icon assets" >&2
    echo "Detected values:" >&2
    echo "  BIN_PATH=${BIN_PATH:-<none>}" >&2
    echo "  DESKTOP_PATH=${DESKTOP_PATH:-<none>}" >&2
    echo "  ICON_PATH=${ICON_PATH:-<none>}" >&2
    exit 1
fi

sed -i -E "s|^Exec=.*|Exec=/usr/bin/${APP_NAME}|" "${DESKTOP_PATH}"
sed -i -E "s|^Icon=.*|Icon=${APP_ID}|" "${DESKTOP_PATH}"

if grep -q '^NoDisplay=' "${DESKTOP_PATH}"; then
    sed -i -E 's|^NoDisplay=.*|NoDisplay=false|' "${DESKTOP_PATH}"
else
    echo "NoDisplay=false" >> "${DESKTOP_PATH}"
fi

if grep -q '^Hidden=' "${DESKTOP_PATH}"; then
    sed -i -E 's|^Hidden=.*|Hidden=false|' "${DESKTOP_PATH}"
fi

if ! grep -q '^Categories=' "${DESKTOP_PATH}"; then
    echo "Categories=Utility;" >> "${DESKTOP_PATH}"
fi

# ---------------------------------------------------------------------
# Debian package from official payload
# ---------------------------------------------------------------------
echo "📦 Creating Debian package..."
mkdir -p "${DEB_ROOT}/DEBIAN"
mkdir -p "${DEB_ROOT}/usr"
cp -a "${USR_NORMALIZED}/." "${DEB_ROOT}/usr/"

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
mkdir -p "${APPDIR}/usr"
cp -a "${USR_NORMALIZED}/." "${APPDIR}/usr/"

cat > "${APPDIR}/AppRun" <<EOF
#!/bin/sh
HERE="\$(dirname "\$(readlink -f "\$0")")"
BIN_NAME="${APP_NAME}"
exec "\$HERE/usr/bin/\$BIN_NAME" "\$@"
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
