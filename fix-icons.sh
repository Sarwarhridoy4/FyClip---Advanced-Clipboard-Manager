#!/usr/bin/env bash
# =====================================================================
# FyClip Icon Fix (system remediation)
# Permanent fix is applied at build time (see build.sh / debian/rules).
# This script remediates an already-installed system where icons are
# missing or mis-sized. It auto-elevates to root via sudo when needed.
# =====================================================================

set -euo pipefail

# Re-exec with sudo if not already root (so it works when called from build.sh)
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
SRC="${ICON_ROOT}/256x256/apps/${APP_ID}.png"
[ -f "${SRC}" ] || SRC="/usr/share/pixmaps/${APP_ID}.png"

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
