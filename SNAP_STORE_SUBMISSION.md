# FyClip - Snap Store Submission Guide

This guide will help you build and submit FyClip to the Ubuntu Snap Store.

## Prerequisites

1. **Install Snapcraft** (if not already installed):
   ```bash
   sudo snap install snapcraft --classic
   ```

2. **Install LXD** (required for building snaps):
   ```bash
   sudo snap install lxd
   sudo lxd init --auto
   sudo usermod -a -G lxd $USER
   newgrp lxd
   ```

3. **Create a Snapcraft account**:
   - Visit https://snapcraft.io and create an account
   - Login to Snapcraft:
     ```bash
     snapcraft login
     ```

## Building the Snap

1. **Build the snap package**:
   ```bash
   snapcraft
   ```

   This will create a `.snap` file in the current directory (e.g., `fyclip_2.2.0_amd64.snap`).

2. **Test the snap locally** (optional but recommended):
   ```bash
   sudo snap install ./fyclip_2.2.0_amd64.snap --dangerous
   fyclip
   ```

   To remove the test installation:
   ```bash
   sudo snap remove fyclip
   ```

## Registering Your App Name

Before publishing, you need to register the app name on the Snap Store:

```bash
snapcraft register fyclip
```

## Publishing to the Snap Store

1. **Upload and publish to the edge channel** (for testing):
   ```bash
   snapcraft upload ./fyclip_2.2.0_amd64.snap --release=edge
   ```

2. **Promote to stable channel** (after testing):
   ```bash
   snapcraft promote fyclip --from-channel=edge --to-channel=stable
   ```

## Snap Store Channels

- **edge**: For development and testing
- **beta**: For beta testers
- **candidate**: For release candidates
- **stable**: For general users

## Updating Your Snap

1. Update the version in `FyneApp.toml` and `snap/snapcraft.yaml`
2. Build the new snap:
   ```bash
   snapcraft
   ```
3. Upload and release:
   ```bash
   snapcraft upload ./fyclip_<version>_amd64.snap --release=stable
   ```

## Snap Configuration

The snap is configured with the following features:

- **Confinement**: Strict (secure sandboxing)
- **Base**: core22 (Ubuntu 22.04)
- **Interfaces**:
  - `desktop`, `desktop-legacy`: Desktop integration
  - `gsettings`: System settings access
  - `home`: Access to user's home directory
  - `network`, `network-bind`: Network access
  - `opengl`: Graphics acceleration
  - `wayland`, `x11`, `unity7`: Display server support
  - `browser-support`: Browser integration
  - `audio-playback`, `pulseaudio`: Audio support

## Troubleshooting

### Build fails with missing dependencies
If the build fails due to missing dependencies, ensure all build-packages are correctly specified in `snapcraft.yaml`.

### App doesn't start
Check the snap logs:
```bash
snap logs fyclip -f
```

### Permission issues
The snap uses strict confinement. If you need additional permissions, update the `plugs` section in `snapcraft.yaml`.

### Icon not showing on Snap Store
If the icon is not displaying on the Snap Store listing:

1. **Verify icon field in snapcraft.yaml**:
   The `icon` field must be present in `snap/snapcraft.yaml`:
   ```yaml
   icon: snap/local/icon.png
   ```

2. **Rebuild and re-upload**:
   ```bash
   snapcraft clean
   snapcraft
   snapcraft upload ./fyclip_2.2.0_amd64.snap --release=edge
   ```

3. **Wait for processing**:
   The icon may take a few minutes to appear after uploading. Refresh the page if needed.

### Links not showing on Snap Store
If the links section shows "No links have been added for this snap":

1. **Verify metadata fields in snapcraft.yaml**:
   Ensure these fields are present in `snap/snapcraft.yaml`:
   ```yaml
   contact: your-email@example.com
   issues: https://github.com/yourusername/yourrepo/issues
   source-code: https://github.com/yourusername/yourrepo
   website: https://yourwebsite.com
   donation: https://yourdonationlink.com
   license: MIT
   ```

2. **Rebuild and re-upload**:
   ```bash
   snapcraft clean
   snapcraft
   snapcraft upload ./fyclip_2.2.0_amd64.snap --release=edge
   ```

3. **Promote to stable** (after verifying):
   ```bash
   snapcraft promote fyclip --from-channel=edge --to-channel=stable
   ```

### Screenshots not showing on Snap Store
If screenshots are not displaying on the Snap Store listing:

1. **Verify screenshots are included in snapcraft.yaml**:
   The screenshots must be organized and primed in the `desktop` part:
   ```yaml
   desktop:
     plugin: dump
     source: snap/local
     organize:
       screenshot1.png: share/fyclip/screenshots/screenshot1.png
       screenshot2.png: share/fyclip/screenshots/screenshot2.png
       screenshot3.png: share/fyclip/screenshots/screenshot3.png
       screenshot4.png: share/fyclip/screenshots/screenshot4.png
       screenshot5.png: share/fyclip/screenshots/screenshot5.png
     prime:
       - share/fyclip/screenshots/*.png
   ```

2. **Verify screenshots exist in snap/local directory**:
   ```bash
   ls -la snap/local/screenshot*.png
   ```

3. **Rebuild and re-upload**:
   ```bash
   snapcraft clean
   snapcraft
   snapcraft upload ./fyclip_2.2.0_amd64.snap --release=edge
   ```

4. **Wait for processing**:
   Screenshots may take a few minutes to appear after uploading. Refresh the page if needed.

### General troubleshooting steps
If you encounter issues after uploading:

1. **Clean and rebuild**:
   ```bash
   snapcraft clean
   snapcraft
   ```

2. **Upload new version**:
   ```bash
   snapcraft upload ./fyclip_2.2.0_amd64.snap --release=edge
   ```

3. **Check snap status**:
   ```bash
   snapcraft status fyclip
   snapcraft list-revisions fyclip
   ```

4. **View snap on store**:
   Visit https://snapcraft.io/fyclip to see your snap listing

## Files Created

- `snap/snapcraft.yaml`: Main snap configuration with metadata
- `snap/local/fyclip.desktop`: Desktop entry file
- `snap/local/icon.png`: Application icon
- `snap/local/screenshot1.png`: Application screenshot 1
- `snap/local/screenshot2.png`: Application screenshot 2
- `snap/local/screenshot3.png`: Application screenshot 3
- `snap/local/screenshot4.png`: Application screenshot 4
- `snap/local/screenshot5.png`: Application screenshot 5

## Additional Resources

- [Snapcraft Documentation](https://snapcraft.io/docs)
- [Snap Store](https://snapcraft.io)
- [Fyne Documentation](https://developer.fyne.io)

## Support

For issues or questions, please visit:
- GitHub: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager
