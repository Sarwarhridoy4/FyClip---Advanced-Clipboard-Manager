# FyClip Distribution Channels

This document summarizes all available distribution channels for FyClip.

## Available Channels

### 1. Snap Store ✅ Available

**Status**: Already published

**Installation**:
```bash
sudo snap install fyclip
```

**Details**:
- Package: `fyclip`
- Version: 2.2.0
- Confinement: strict
- Architecture: amd64, arm64

**Links**:
- Snapcraft: https://snapcraft.io/fyclip
- GitHub: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager

---

### 2. Debian Package ✅ Available

**Status**: Package built and ready

**Installation**:
```bash
# Download from GitHub Releases
wget https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/releases/download/v2.2.0/fyclip_2.2.0-1_amd64.deb

# Install
sudo dpkg -i fyclip_2.2.0-1_amd64.deb
sudo apt-get install -f
```

**Build**:
```bash
./build-deb.sh
# or
make deb
```

**Details**:
- Package: `fyclip`
- Version: 2.2.0-1
- Architecture: amd64
- Section: utils

---

### 3. Ubuntu PPA 🔄 Ready to Submit

**Status**: Source package ready for upload

**Setup**:
1. Create PPA on Launchpad: https://launchpad.net/~YOUR_USERNAME
2. Upload source package:
   ```bash
   ./build-ppa.sh
   dput ppa:YOUR_USERNAME/fyclip dist/fyclip_2.2.0-1_source.changes
   ```

**Installation** (after PPA is set up):
```bash
sudo add-apt-repository ppa:YOUR_USERNAME/fyclip
sudo apt-get update
sudo apt-get install fyclip
```

**Build**:
```bash
./build-ppa.sh
# or
make ppa
```

**Details**:
- Package: `fyclip`
- Version: 2.2.0-1
- Architecture: amd64
- PPA: ppa:YOUR_USERNAME/fyclip

---

### 4. Debian Official Repository 🔄 To Be Submitted

**Status**: Ready for submission to Debian

**Process**:
1. Create Salsa account: https://salsa.debian.org
2. Push code to Salsa repository
3. Request sponsorship via debian-mentors
4. Package reviewed and approved
5. Available in Debian unstable
6. Migrates to testing
7. Included in next stable release

**Installation** (after approval):
```bash
sudo apt-get install fyclip
```

**Documentation**:
- See [SALSA_SUBMISSION.md](SALSA_SUBMISSION.md) for detailed guide

---

### 5. Ubuntu Official Repository 🔄 Pending Debian

**Status**: Automatic after Debian approval

**Process**:
1. Package enters Debian unstable
2. Automatically synced to Ubuntu
3. Available in Ubuntu App Center

**Installation** (after sync):
```bash
sudo apt-get install fyclip
```

---

### 6. Flatpak via Flathub 🔄 To Be Created

**Status**: Manifest ready for submission

**Setup**:
1. Fork Flathub repository: https://github.com/flathub/flathub
2. Add Flatpak manifest
3. Submit pull request
4. Wait for review and approval

**Installation** (after approval):
```bash
flatpak install flathub com.sarwar.fyclip
```

**Build** (local):
```bash
flatpak-builder build-dir com.sarwar.fyclip.yml --force-clean
flatpak-builder --user --install build-dir com.sarwar.fyclip.yml
```

---

## Quick Reference

| Channel | Status | Install Command | Build Command |
|---------|--------|-----------------|---------------|
| Snap Store | ✅ Available | `sudo snap install fyclip` | N/A |
| Debian Package | ✅ Available | `sudo dpkg -i fyclip_2.2.0-1_amd64.deb` | `./build-deb.sh` |
| Ubuntu PPA | 🔄 Ready | `sudo add-apt-repository ppa:USER/fyclip` | `./build-ppa.sh` |
| Debian Official | 🔄 To Submit | `sudo apt-get install fyclip` | N/A |
| Ubuntu Official | 🔄 Pending | `sudo apt-get install fyclip` | N/A |
| Flatpak | 🔄 To Create | `flatpak install flathub com.sarwar.fyclip` | `flatpak-builder` |

## Distribution Strategy

### Immediate Availability
1. **Snap Store** - Already available
2. **GitHub Releases** - Debian package available
3. **Ubuntu PPA** - Ready to create

### Short-term (1-3 months)
1. **Submit to Debian** - Official repository
2. **Submit to Flathub** - Wide distribution

### Long-term (6-12 months)
1. **Debian approval** - Package in Debian unstable
2. **Ubuntu sync** - Automatic inclusion in Ubuntu
3. **Official App Center** - Available in Ubuntu Software

## Build Commands Summary

```bash
# Build Debian package
./build-deb.sh
make deb

# Build PPA source package
./build-ppa.sh
make ppa

# Build with specific version
./build-deb.sh 2.2.0
./build-ppa.sh 2.2.0
make deb-version VERSION=2.2.0
make ppa-version VERSION=2.2.0
```

## Documentation

- [DEBIAN_PACKAGING.md](DEBIAN_PACKAGING.md) - Debian packaging guide
- [SALSA_SUBMISSION.md](SALSA_SUBMISSION.md) - Debian Salsa submission guide
- [APP_CENTER_SUBMISSION.md](APP_CENTER_SUBMISSION.md) - Ubuntu App Center guide
- [DEBIAN_PACKAGING_SUMMARY.md](DEBIAN_PACKAGING_SUMMARY.md) - Summary of all files

## Resources

- [Snapcraft](https://snapcraft.io)
- [Launchpad](https://launchpad.net)
- [Debian Salsa](https://salsa.debian.org)
- [Flathub](https://flathub.org)
- [Ubuntu App Center](https://apps.ubuntu.com)

## Contact

For questions about distribution:
- **Maintainer**: Sarwar Hossain <sarwarhridoy4@gmail.com>
- **Upstream**: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager

## License

FyClip is licensed under the MIT License. See [Licence](Licence) for details.
