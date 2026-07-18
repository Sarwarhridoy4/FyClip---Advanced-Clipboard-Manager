# FyClip Icon Fix Issue

## Problem

The FyClip application icon (`com.sarwar.fyclip`) has size mismatches in the system icon theme, causing display issues in desktop environments.

### Symptoms
- FyClip icon may appear blurry or incorrectly scaled in panels, menus, and file managers
- Icon fallback to generic or missing icon in some contexts

### Root Cause

The icon files installed in `/usr/share/icons/hicolor/` do not match their directory names:

| Location | Actual Size | Expected Size |
|----------|-------------|---------------|
| `128x128/apps/com.sarwar.fyclip.png` | 512x512 | 128x128 |
| `256x256/apps/com.sarwar.fyclip.png` | 512x512 | 256x256 |

Additionally, several standard icon sizes are missing from the icon theme:
- 16x16, 24x24, 32x32, 48x48, 64x64

And the `StartupWMClass` in the desktop entry does not match the actual window class returned by `xprop` (`FyClip - Clipboard Manager` vs `FyClip Clipboard Manager`), which can prevent the window manager from matching the icon to the application.

## Fix

### 1. Fix StartupWMClass mismatch
```bash
sudo sed -i 's/^StartupWMClass=.*/StartupWMClass=FyClip - Clipboard Manager/' /usr/share/applications/com.sarwar.fyclip.desktop
```

### 2. Create missing icon sizes
```bash
sudo mkdir -p /usr/share/icons/hicolor/{16x16,24x24,32x32,48x48,64x64}/apps
sudo python3 -c "
from PIL import Image
src = '/usr/share/pixmaps/com.sarwar.fyclip.png'
img = Image.open(src).convert('RGBA')
for s in [16, 24, 32, 48, 64]:
    out = f'/usr/share/icons/hicolor/{s}x{s}/apps/com.sarwar.fyclip.png'
    resized = img.resize((s, s), Image.Resampling.LANCZOS)
    resized.save(out)
"
```

### 3. Update system caches
```bash
sudo gtk-update-icon-cache /usr/share/icons/hicolor
sudo update-desktop-database /usr/share/applications
```

## Verification

Check that icon files now have correct dimensions:
```bash
file /usr/share/icons/hicolor/*/apps/com.sarwar.fyclip.png
```

Expected output should show matching sizes (e.g., `128 x 128` for files in `128x128/` directory).
