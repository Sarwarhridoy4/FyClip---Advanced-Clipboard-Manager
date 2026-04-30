# FyClip Debian Packaging

This document describes how to build and submit FyClip as a Debian package.

## Prerequisites

Install the required build dependencies:

```bash
sudo apt-get install build-essential debhelper dpkg-dev golang-go
```

## Building the Debian Package

### Quick Build

Use the provided build script:

```bash
./build-deb.sh
```

This will:
1. Check for required dependencies
2. Clean previous builds
3. Prepare the source
4. Build the Debian package
5. Verify the package

The resulting `.deb` file will be in the `dist/` directory.

### Using Make

You can also use the Makefile targets:

```bash
# Build Debian package
make deb

# Build with specific version
make deb-version VERSION=2.2.2
```

### Manual Build

If you prefer to build manually:

```bash
# Install build dependencies
sudo apt-get install build-essential debhelper dpkg-dev golang-go

# Build the package
dpkg-buildpackage -us -uc -b
```

## Installing the Package

After building, install the package:

```bash
sudo dpkg -i dist/fyclip_2.2.2-1_amd64.deb
sudo apt-get install -f  # Fix any dependency issues
```

## Removing the Package

```bash
sudo dpkg -r fyclip
```

## Package Structure

The Debian package includes:

- **Binary**: `/usr/bin/fyclip`
- **Desktop file**: `/usr/share/applications/com.sarwar.fyclip.desktop`
- **Icon**: `/usr/share/icons/hicolor/256x256/apps/com.sarwar.fyclip.png`
- **Screenshots**: `/usr/share/fyclip/screenshots/`

## Dependencies

The package depends on:

- `libgl1` - OpenGL library
- `libx11-6` - X11 client-side library
- `libxcursor1` - X cursor management library
- `libxrandr2` - X11 RandR extension library
- `libxinerama1` - X11 Xinerama extension library
- `libxi6` - X11 Input extension library
- `libgtk-3-0` - GTK+ 3 graphical user interface library
- `xclip | xsel | wl-clipboard` - Clipboard utilities

## Submitting to Debian

To submit this package to Debian:

1. **Create a Debian Salsa account**: https://salsa.debian.org

2. **Fork the package repository**: Create a new project on Salsa

3. **Upload your source**: Push your debian/ directory to the repository

4. **Request sponsorship**: Follow the Debian sponsorship process

5. **Lintian checks**: Run lintian to check for issues:
   ```bash
   lintian -i -I --show-overrides dist/fyclip_2.2.2-1_amd64.changes
   ```

## Troubleshooting

### Missing Dependencies

If you get dependency errors:

```bash
sudo apt-get update
sudo apt-get install -f
```

### Build Failures

Check the build log for errors. Common issues:

- Missing Go: Install with `sudo apt-get install golang-go`
- Missing debhelper: Install with `sudo apt-get install debhelper`
- Missing dpkg-dev: Install with `sudo apt-get install dpkg-dev`

### Package Installation Issues

If the package fails to install:

```bash
# Check dependencies
dpkg -I dist/fyclip_2.2.2-1_amd64.deb

# Force installation (not recommended)
sudo dpkg --force-depends -i dist/fyclip_2.2.2-1_amd64.deb
```

## Resources

- [Debian New Maintainer's Guide](https://www.debian.org/doc/manuals/maint-guide/)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Debian Packaging Tutorial](https://www.debian.org/doc/manuals/packaging-tutorial/)
- [Lintian](https://lintian.debian.org/)

## License

FyClip is licensed under the MIT License. See [Licence](Licence) for details.
