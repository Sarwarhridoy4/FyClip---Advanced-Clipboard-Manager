#!/usr/bin/env bash
# =====================================================================
# FyClip Icon Fix (system remediation / post-install helper)
# Permanent fix is applied at build/packaging time (see build.sh /
# debian/rules). This script remediates an already-installed system:
# it regenerates all standard hicolor icon sizes from a source image,
# fixes the desktop StartupWMClass, and refreshes icon caches.
#
# Usage:
#   sudo ./fix-icons.sh [source_icon.png]
#
# When run without root it auto-elevates via sudo. An optional source
# icon path may be supplied (e.g. the freshly built 256x256 icon).
# If omitted, the script searches repo icon.png, the installed hicolor
# 256x256 icon, then /usr/share/pixmaps.
# =====================================================================

set -euo pipefail

# Re-exec with sudo if not already root (standalone use).
if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        exec sudo "$0" "$@"
    else
        echo "This script must be run as root (no sudo available)." >&2
        exit 1
    fi
fi

APP_ID="com.sarwar.fyclip"
ICON_ROOT="/usr/share/icons/hicolor"
DESKTOP="/usr/share/applications/${APP_ID}.desktop"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Resolve source icon. Priority:
#   1. Explicit argument
#   2. Repo icon.png (present when run from the project / tarball root)
#   3. Just-installed hicolor 256x256 (postinst context)
#   4. /usr/share/pixmaps
SRC="${1:-}"
if [ -z "${SRC}" ] || [ ! -f "${SRC}" ]; then
    if [ -f "${SCRIPT_DIR}/icon.png" ]; then
        SRC="${SCRIPT_DIR}/icon.png"
    elif [ -f "${ICON_ROOT}/256x256/apps/${APP_ID}.png" ]; then
        SRC="${ICON_ROOT}/256x256/apps/${APP_ID}.png"
    elif [ -f "/usr/share/pixmaps/${APP_ID}.png" ]; then
        SRC="/usr/share/pixmaps/${APP_ID}.png"
    fi
fi

if [ ! -f "${SRC}" ]; then
    echo "ERROR: could not locate a source icon (tried arg, icon.png, hicolor 256x256, pixmaps)." >&2
    exit 1
fi
echo "Using source icon: ${SRC}"

# 1. Fix StartupWMClass mismatch (ensure it matches xprop's window class)
if [ -f "${DESKTOP}" ]; then
    sed -i 's/^StartupWMClass=.*/StartupWMClass=FyClip - Clipboard Manager/' "${DESKTOP}"
fi

# 2. (Re)generate all standard sizes, properly resized to each directory
python3 - <<PYEOF
from PIL import Image
import os
src = "${SRC}"
img = Image.open(src).convert("RGBA")
for s in [16, 24, 32, 48, 64, 128, 256]:
    d = "${ICON_ROOT}/%dx%d/apps" % (s, s)
    os.makedirs(d, exist_ok=True)
    out = os.path.join(d, "${APP_ID}.png")
    img.resize((s, s), getattr(Image, 'Resampling', Image).LANCZOS).save(out)
print("icons generated")
PYEOF

# 3. Update system caches
gtk-update-icon-cache -f "${ICON_ROOT}" 2>/dev/null || true
update-desktop-database /usr/share/applications 2>/dev/null || true

echo "Done. Verifying icon dimensions:"
file "${ICON_ROOT}"/*/apps/${APP_ID}.png
