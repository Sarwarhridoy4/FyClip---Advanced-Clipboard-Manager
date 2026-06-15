# Ubuntu App Center Submission Guide for FyClip

This guide explains how to submit FyClip to the Ubuntu App Center.

## Overview

The Ubuntu App Center (formerly Ubuntu Software Center) is the default application store for Ubuntu. You can submit your application in multiple ways:

1. **Snap Store** (Recommended) - Already available
2. **Debian Package** - Via Ubuntu repositories
3. **Flatpak** - Via Flathub
4. **PPA** - Personal Package Archive

## Option 1: Snap Store (Already Available)

FyClip is already available as a Snap package. Users can install it via:

```bash
sudo snap install fyclip
```

## Option 2: Ubuntu PPA (Personal Package Archive)

A PPA allows you to distribute your Debian packages directly to Ubuntu users.

### Step 1: Create Launchpad Account

1. Go to https://launchpad.net
2. Create an account or log in
3. Set up your GPG key

### Step 2: Create PPA

1. Go to https://launchpad.net/~YOUR_USERNAME
2. Click "Create a new PPA"
3. Fill in:
   - **Name**: `fyclip`
   - **Display name**: `FyClip Clipboard Manager`
   - **Description**: `Advanced Clipboard Manager for Linux`

### Step 3: Prepare Source Package

```bash
# Build source package
./build-deb.sh

# Create source package for PPA
cd dist
dpkg-source -b ../fyclip-2.2.2
```

### Step 4: Upload to PPA

```bash
# Install dput
sudo apt-get install dput

# Upload to PPA
dput ppa:YOUR_USERNAME/fyclip fyclip_2.2.2-1_source.changes
```

### Step 5: Add PPA to Ubuntu App Center

Users can add your PPA and install FyClip:

```bash
# Add PPA
sudo add-apt-repository ppa:YOUR_USERNAME/fyclip

# Update package list
sudo apt-get update

# Install FyClip
sudo apt-get install fyclip
```

## Option 3: Flatpak via Flathub

Flathub is the largest app store for Linux.

### Step 1: Install Flatpak

```bash
sudo apt-get install flatpak
```

### Step 2: Create Flatpak Manifest

Create `com.sarwar.fyclip.yml`:

```yaml
app-id: com.sarwar.fyclip
runtime: org.freedesktop.Platform
runtime-version: '23.08'
sdk: org.freedesktop.Sdk
sdk-extensions:
  - org.freedesktop.Sdk.Extension.golang
command: fyclip

finish-args:
  - --share=ipc
  - --socket=x11
  - --socket=wayland
  - --socket=fallback-x11
  - --share=network
  - --device=dri
  - --filesystem=home
  - --talk-name=org.freedesktop.Flatpak
  - --talk-name=org.freedesktop.portal.Flatpak

modules:
  - name: fyclip
    buildsystem: simple
    build-commands:
      - go build -o fyclip .
      - install -Dm755 fyclip /app/bin/fyclip
      - install -Dm644 icon.png /app/share/icons/hicolor/256x256/apps/com.sarwar.fyclip.png
      - install -Dm644 com.sarwar.fyclip.desktop /app/share/applications/com.sarwar.fyclip.desktop
    sources:
      - type: git
        url: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager.git
        tag: v2.2.2
```

### Step 3: Build Flatpak

```bash
# Install flatpak-builder
sudo apt-get install flatpak-builder

# Build Flatpak
flatpak-builder build-dir com.sarwar.fyclip.yml --force-clean

# Install locally
flatpak-builder --user --install build-dir com.sarwar.fyclip.yml
```

### Step 4: Submit to Flathub

1. Fork https://github.com/flathub/flathub
2. Add your manifest to the repository
3. Submit a pull request
4. Wait for review and approval

## Option 4: Direct Submission to Ubuntu

For official Ubuntu repository inclusion:

### Requirements

1. **Debian Developer**: You must be a Debian Developer or have a sponsor
2. **Package Quality**: Package must pass all lintian checks
3. **Policy Compliance**: Must follow Debian Policy Manual
4. **Documentation**: Complete manpages and documentation

### Process

1. Submit to Debian first (see SALSA_SUBMISSION.md)
2. Package migrates from Debian unstable to testing
3. Package is synced to Ubuntu repositories
4. Available in Ubuntu App Center

## Recommended Approach

For immediate availability in Ubuntu App Center:

1. **Use Snap Store** (Already done)
2. **Create PPA** (Quick and easy)
3. **Submit to Flathub** (Wide distribution)

For long-term official support:

1. **Submit to Debian** (Official repositories)
2. **Automatic sync to Ubuntu** (Official App Center)

## Quick Start: Create PPA

```bash
# 1. Build source package
./build-deb.sh

# 2. Create PPA on Launchpad
# Go to https://launchpad.net/~YOUR_USERNAME

# 3. Upload to PPA
dput ppa:YOUR_USERNAME/fyclip dist/fyclip_2.2.2-1_source.changes

# 4. Share with users
# Users can install with:
# sudo add-apt-repository ppa:YOUR_USERNAME/fyclip
# sudo apt-get update
# sudo apt-get install fyclip
```

## Your PPA

**PPA URL**: https://launchpad.net/~sarwar-hossain/+archive/ubuntu/fyclip

**Installation Instructions**:
```bash
sudo add-apt-repository ppa:sarwar-hossain/fyclip
sudo apt-get update
sudo apt-get install fyclip
```

## Distribution Channels

| Channel | Status | Installation |
|---------|--------|--------------|
| Snap Store | ✅ Available | `sudo snap install fyclip` |
| Ubuntu PPA | 🔄 To be created | `sudo add-apt-repository ppa:YOUR_USERNAME/fyclip` |
| Flathub | 🔄 To be created | `flatpak install flathub com.sarwar.fyclip` |
| Debian | 🔄 To be submitted | `sudo apt-get install fyclip` |
| Ubuntu Official | 🔄 Pending Debian | `sudo apt-get install fyclip` |

## Resources

- [Launchpad](https://launchpad.net)
- [PPA Documentation](https://help.launchpad.net/Packaging/PPA)
- [Flathub](https://flathub.org)
- [Flatpak Documentation](https://docs.flatpak.org)
- [Ubuntu App Center](https://apps.ubuntu.com)

## Contact

For questions about App Center submission:
- **Maintainer**: Sarwar Hossain <sarwarhridoy4@gmail.com>
- **Upstream**: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager

## License

FyClip is licensed under the MIT License. See [Licence](Licence) for details.
