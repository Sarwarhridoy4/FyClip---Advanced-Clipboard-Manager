# Debian Packaging Summary for FyClip

This document summarizes all files created for Debian packaging.

## Files Created

### Debian Packaging Files

| File | Description |
|------|-------------|
| [`debian/control`](debian/control) | Package metadata and dependencies |
| [`debian/rules`](debian/rules) | Build instructions using debhelper |
| [`debian/changelog`](debian/changelog) | Version history |
| [`debian/copyright`](debian/copyright) | MIT license information |
| [`debian/source/format`](debian/source/format) | Source format (3.0 native) |
| [`debian/fyclip.desktop`](debian/fyclip.desktop) | Desktop entry file |
| [`debian/fyclip.1`](debian/fyclip.1) | Manpage for fyclip command |
| [`debian/postinst`](debian/postinst) | Post-installation script |
| [`debian/postrm`](debian/postrm) | Post-removal script |
| [`debian/watch`](debian/watch) | Upstream version tracking |

### Build Scripts

| File | Description |
|------|-------------|
| [`build-deb.sh`](build-deb.sh) | Automated Debian package builder |
| [`Makefile`](Makefile) | Updated with `make deb` and `make deb-version` targets |

### Documentation

| File | Description |
|------|-------------|
| [`DEBIAN_PACKAGING.md`](DEBIAN_PACKAGING.md) | Comprehensive guide for building Debian packages |
| [`SALSA_SUBMISSION.md`](SALSA_SUBMISSION.md) | Guide for submitting to Debian Salsa |
| [`DEBIAN_PACKAGING_SUMMARY.md`](DEBIAN_PACKAGING_SUMMARY.md) | This file |

### GitHub Actions

| File | Description |
|------|-------------|
| [`.github/workflows/build-deb.yml`](.github/workflows/build-deb.yml) | Automated Debian package building workflow |

## Package Details

- **Package Name**: fyclip
- **Version**: 2.2.2-1
- **Architecture**: amd64
- **Section**: utils
- **Priority**: optional
- **Maintainer**: Sarwar Hossain <sarwarhridoy4@gmail.com>
- **Homepage**: https://fyclip.vercel.app
- **License**: MIT

## Dependencies

### Build Dependencies
- debhelper-compat (= 13)
- libgl1-mesa-dev
- libxcursor-dev
- libxrandr-dev
- libxinerama-dev
- libxi-dev
- libgl-dev
- libxxf86vm-dev
- libxkbcommon-dev
- libwayland-dev
- libxkbcommon-x11-dev
- libx11-dev
- libxext-dev
- libxrender-dev
- libxfixes-dev
- libxss-dev
- libxtst-dev
- libasound2-dev
- libpulse-dev
- golang-go (>= 1.21)

### Runtime Dependencies
- libgl1
- libx11-6
- libxcursor1
- libxrandr2
- libxinerama1
- libxi6
- libgtk-3-0
- xclip | xsel | wl-clipboard

## Package Contents

```
/usr/bin/fyclip
/usr/share/applications/com.sarwar.fyclip.desktop
/usr/share/icons/hicolor/256x256/apps/com.sarwar.fyclip.png
/usr/share/fyclip/screenshots/screenshot1.png
/usr/share/fyclip/screenshots/screenshot2.png
/usr/share/fyclip/screenshots/screenshot3.png
/usr/share/fyclip/screenshots/screenshot4.png
/usr/share/fyclip/screenshots/screenshot5.png
/usr/share/man/man1/fyclip.1.gz
/usr/share/doc/fyclip/changelog.Debian.gz
/usr/share/doc/fyclip/copyright
```

## Build Commands

### Quick Build
```bash
./build-deb.sh
```

### Using Make
```bash
make deb
make deb-version VERSION=2.2.2
```

### Manual Build
```bash
dpkg-buildpackage -us -uc -b
```

## Installation

```bash
sudo dpkg -i dist/fyclip_2.2.2-1_amd64.deb
sudo apt-get install -f
```

## Removal

```bash
sudo dpkg -r fyclip
```

## Verification

```bash
# Check package info
dpkg-deb --info dist/fyclip_2.2.2-1_amd64.deb

# Check package contents
dpkg-deb --contents dist/fyclip_2.2.2-1_amd64.deb

# Run lintian checks
lintian -i -I --show-overrides dist/fyclip_2.2.2-1_amd64.changes
```

## Submission to Debian

See [`SALSA_SUBMISSION.md`](SALSA_SUBMISSION.md) for detailed instructions on submitting to Debian Salsa.

## GitHub Actions

The [`.github/workflows/build-deb.yml`](.github/workflows/build-deb.yml) workflow automatically:
- Builds Debian package on push/PR to main/master
- Creates packages for tagged releases
- Uploads packages as artifacts
- Publishes to GitHub Releases on tag push

## Resources

- [Debian New Maintainer's Guide](https://www.debian.org/doc/manuals/maint-guide/)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Debian Packaging Tutorial](https://www.debian.org/doc/manuals/packaging-tutorial/)
- [Lintian](https://lintian.debian.org/)
- [Debian Mentors](https://mentors.debian.net/)
- [Debian Salsa](https://salsa.debian.org/)

## License

FyClip is licensed under the MIT License. See [Licence](Licence) for details.
