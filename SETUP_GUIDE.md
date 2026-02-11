# FyClip Setup Guide

## Complete File Structure

Create the following directory structure:

```
fyclip/
├── main.go
├── go.mod
├── icon.png                    (your application icon)
├── README.md
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── clipboard/
│   │   ├── item.go
│   │   ├── manager.go
│   │   ├── monitor.go
│   │   ├── native.go
│   │   └── storage.go
│   ├── ui/
│   │   ├── window.go
│   │   ├── list.go
│   │   ├── preview.go
│   │   ├── toolbar.go
│   │   ├── search.go
│   │   ├── status.go
│   │   └── dialogs.go
│   ├── platform/
│   │   ├── autostart.go
│   │   ├── autostart_linux.go
│   │   ├── autostart_windows.go
│   │   └── autostart_darwin.go
│   └── tray/
│       └── tray.go
```

## Quick Setup Steps

### 1. Create Project Directory

```bash
mkdir -p fyclip
cd fyclip
```

### 2. Create All Subdirectories

```bash
mkdir -p internal/app
mkdir -p internal/clipboard
mkdir -p internal/ui
mkdir -p internal/platform
mkdir -p internal/tray
```

### 3. Initialize Go Module

```bash
go mod init fyclip
```

### 4. Copy All Source Files

Copy each `.go` file to its corresponding directory as shown in the structure above.

### 5. Add Your Icon

Place your `icon.png` file in the root directory. If you don't have one, you can:

```bash
# Download a placeholder icon or create one
# The icon should be PNG format, 512x512 pixels recommended
```

### 6. Install Dependencies

```bash
go mod tidy
```

This will automatically download project dependencies from `go.mod`.

### 7. Build the Application

```bash
go build -o fyclip
```

For Windows:
```bash
go build -ldflags="-H=windowsgui" -o fyclip.exe
```

### 8. Run

```bash
./fyclip
```

## File Creation Checklist

Use this checklist to ensure all files are created:

- [ ] `main.go` - Application entry point
- [ ] `go.mod` - Go module file
- [ ] `icon.png` - Application icon
- [ ] `README.md` - Documentation
- [ ] `internal/app/app.go` - App initialization
- [ ] `internal/clipboard/item.go` - Item types
- [ ] `internal/clipboard/manager.go` - Manager logic
- [ ] `internal/clipboard/monitor.go` - Clipboard monitoring
- [ ] `internal/clipboard/native.go` - Native clipboard
- [ ] `internal/clipboard/storage.go` - Persistence
- [ ] `internal/ui/window.go` - Main window
- [ ] `internal/ui/list.go` - History list
- [ ] `internal/ui/preview.go` - Preview pane
- [ ] `internal/ui/toolbar.go` - Toolbar
- [ ] `internal/ui/search.go` - Search bar
- [ ] `internal/ui/status.go` - Status bar
- [ ] `internal/ui/dialogs.go` - Dialogs
- [ ] `internal/platform/autostart.go` - Autostart interface
- [ ] `internal/platform/autostart_linux.go` - Linux autostart
- [ ] `internal/platform/autostart_windows.go` - Windows autostart
- [ ] `internal/platform/autostart_darwin.go` - macOS autostart
- [ ] `internal/tray/tray.go` - System tray

## Verifying Installation

After setup, verify everything works:

```bash
# Check all imports
go mod verify

# Run tests (if you add them)
go test ./...

# Build
go build

# Run
./fyclip
```

## Common Issues

### Issue: Import errors

**Solution**: Run `go mod tidy` to download dependencies

### Issue: Icon not loading

**Solution**: Make sure `icon.png` exists in the root directory and the embed path in `app.go` is correct

### Issue: Clipboard not working on Linux

**Solution**: Install clipboard tools:
```bash
# For X11
sudo apt install xclip

# For Wayland
sudo apt install wl-clipboard
```

### Issue: Build fails on Windows

**Solution**: Use the correct build flags:
```bash
go build -ldflags="-H=windowsgui" -o fyclip.exe
```

## Development Workflow

### Making Changes

1. Edit the relevant files
2. Run `go build` to test compilation
3. Test the application
4. Commit changes

### Adding New Features

1. Identify the appropriate module
2. Add functionality to that module
3. Update UI if needed
4. Test thoroughly
5. Update documentation

### Code Organization

- **Business Logic**: `internal/clipboard/`
- **User Interface**: `internal/ui/`
- **Platform Code**: `internal/platform/`
- **System Integration**: `internal/tray/`
- **Application Setup**: `internal/app/`

## Performance Tips

The modular architecture provides several performance benefits:

1. **Debounced Updates**: UI refreshes are batched
2. **Async Storage**: File I/O doesn't block UI
3. **Efficient Locking**: Minimal lock contention
4. **Smart Filtering**: Early returns for better search

## Testing Your Build

### Basic Functionality Test

1. **Launch Application**: Should open main window
2. **Copy Text**: Copy something, check it appears in history
3. **Copy Image**: Take a screenshot, should appear in history
4. **Search**: Type in search bar, list should filter
5. **Pin Item**: Click pin button, item should stay at top
6. **Delete Item**: Select and delete, item should disappear
7. **Preview**: Select item, preview should update
8. **Copy from History**: Select item, click copy, paste elsewhere
9. **System Tray**: Close window, should minimize to tray
10. **AutoStart**: Enable in tray menu, restart computer to verify

## Building for Distribution

### Linux

```bash
# Recommended: use the project build script
./build.sh

# Optional explicit version
./build.sh 1.5.1
```

This script now follows the official Fyne Linux packaging flow (`fyne package --os linux`) and then generates:
- `dist/fyclip_<version>_<arch>.deb`
- `dist/fyclip_<version>_<arch>.AppImage`

Required tools:
- `go`
- `fyne` CLI
- `dpkg-deb`
- `appimagetool`
- `tar`

### Windows

```bash
# Build with hidden console
go build -ldflags="-s -w -H=windowsgui" -o fyclip.exe

# Create distribution folder
mkdir fyclip-windows-amd64
copy fyclip.exe fyclip-windows-amd64\
copy icon.png fyclip-windows-amd64\
copy README.md fyclip-windows-amd64\

# Create zip archive
# (Use 7-Zip or similar)
```

### macOS

```bash
# Build universal binary
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o fyclip-amd64
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o fyclip-arm64
lipo -create -output fyclip fyclip-amd64 fyclip-arm64

# Create app bundle (optional)
# See Fyne documentation for creating .app bundles
```

## Next Steps

After setup:

1. **Test thoroughly** on your target platforms
2. **Add custom features** as needed
3. **Create tests** for critical functionality
4. **Package for distribution**
5. **Set up CI/CD** for automated builds

## Support

If you encounter issues:

1. Check the README.md troubleshooting section
2. Verify all files are in the correct locations
3. Ensure dependencies are installed
4. Check Go version (1.21+)
5. Review build logs for errors

## Resources

- [Fyne Documentation](https://docs.fyne.io/)
- [Fyne Packaging](https://docs.fyne.io/started/packaging.html)
- [Fyne Metadata](https://docs.fyne.io/started/metadata)
- [Go Documentation](https://go.dev/doc/)
- [Project Repository](https://github.com/Sarwarhridoy4/fyclip)

---
