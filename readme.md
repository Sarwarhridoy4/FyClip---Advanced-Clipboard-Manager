# FyClip - Advanced Clipboard Manager

A modular, high-performance clipboard manager built with Go and Fyne v2.7+.

**Current Version**: 1.5.0

## Features

- 📋 **Clipboard History**: Automatically saves text and images
- 📌 **Pin Items**: Keep important items at the top
- ⭐ **Favorites View**: Toggle pinned-only view instantly
- 🔍 **Search**: Quick search through clipboard history
- ❌ **Clear Search**: One-click reset for the search box
- 🖼️ **Image Support**: Preview and save clipboard images
- 📤 **Unified Export**: Export selected text or images from one action
- 📝 **Markdown Preview**: Markdown content renders correctly in preview pane
- 🕒 **Relative Time + Reuse Count**: List rows show recency and copy frequency
- 💾 **Persistent Storage**: History saved across sessions
- 🚀 **AutoStart**: Launch on system startup
- ⏸️ **Pause Capture**: Pause monitoring for 5 minutes from toolbar/tray
- 🎨 **Modern UI**: Dark theme with responsive design
- ⚡ **Performance**: Debounced updates, async operations
- 🐧 **Linux Packaging**: Official Fyne Linux package pipeline for `.deb` and `.AppImage`
- 🔒 **Thread-Safe**: Proper concurrency handling

## Improvements

### Recently Implemented

- Fixed pin-toggle behavior from list items to save state without shutting down clipboard monitoring.
- Added a debounced, serialized history save pipeline to reduce frequent disk writes during rapid copy events.
- Fixed programmatic image-copy suppression by hashing raw image bytes for correct deduplication behavior.

### Planned Enhancements

- Add global hotkey quick panel for fast paste from recent history.
- Add snippets/templates with titles and categories.
- Add app/pattern exclusion rules and temporary pause mode for sensitive workflows.
- Improve search speed further with indexed/lowercased cache paths.
- Add encrypted import/export backup support.

## Project Structure

```
fyclip/
├── main.go                      # Application entry point
├── go.mod                       # Go module dependencies
├── icon.png                     # Application icon
├── internal/
│   ├── app/
│   │   └── app.go              # App initialization
│   ├── clipboard/
│   │   ├── item.go             # Clipboard item types
│   │   ├── manager.go          # Core manager logic
│   │   ├── monitor.go          # Clipboard monitoring
│   │   ├── native.go           # Platform clipboard ops
│   │   └── storage.go          # Persistence layer
│   ├── ui/
│   │   ├── window.go           # Main window
│   │   ├── list.go             # History list
│   │   ├── preview.go          # Preview pane
│   │   ├── toolbar.go          # Action buttons
│   │   ├── search.go           # Search bar
│   │   ├── status.go           # Status bar
│   │   └── dialogs.go          # Dialogs & utilities
│   ├── platform/
│   │   ├── autostart.go        # Autostart interface
│   │   ├── autostart_linux.go  # Linux implementation
│   │   ├── autostart_windows.go # Windows implementation
│   │   └── autostart_darwin.go # macOS implementation
│   └── tray/
│       └── tray.go             # System tray
└── README.md
```

## Requirements

- Go 1.21 or later
- Fyne v2.5+
- For Linux:
  - X11: `xclip` package
  - Wayland: `wl-clipboard` package
- For Windows: No additional dependencies
- For macOS: No additional dependencies

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager.git
cd FyClip---Advanced-Clipboard-Manager
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Build

```bash
go build -o fyclip
```

### 4. Run

```bash
./fyclip
```

## Linux Dependencies

### Ubuntu/Debian

```bash
# For X11
sudo apt install xclip

# For Wayland
sudo apt install wl-clipboard
```

### Arch Linux

```bash
# For X11
sudo pacman -S xclip

# For Wayland
sudo pacman -S wl-clipboard
```

### Fedora

```bash
# For X11
sudo dnf install xclip

# For Wayland
sudo dnf install wl-clipboard
```

## Building for Release

### Linux Debian + AppImage (Recommended)

Use the project script, which now follows Fyne's official Linux packaging flow (`fyne package --os linux`) and then builds:
- Debian package: `dist/fyclip_<version>_<arch>.deb`
- AppImage: `dist/fyclip_<version>_<arch>.AppImage`

```bash
./build.sh
# or pass explicit version
./build.sh 1.5.1
```

Requirements for the script:
- `go`
- `fyne` CLI
- `dpkg-deb`
- `appimagetool`
- `tar`

Packaging process used by `build.sh`:
1. Run Fyne official Linux packaging:
   - `fyne package --os linux --release --name fyclip --icon icon.png`
   - Generates `fyclip.tar.xz`
2. Extract Fyne package payload and reuse its generated Linux assets:
   - Binary in `usr/bin` or `usr/local/bin`
   - Desktop entry in `usr/share/applications` or `usr/local/share/applications`
   - Icon in `usr/share/pixmaps` or `usr/local/share/pixmaps`
3. Build Debian package (`.deb`) from that payload via `dpkg-deb`
4. Build AppImage (`.AppImage`) from the same payload via `appimagetool`
5. Place final artifacts in `dist/`

### Fyne Native Packaging

Package using Fyne directly:

```bash
# Linux tar package (official Fyne output)
fyne package --os linux --release --name fyclip --icon icon.png

# Windows installer
fyne package --os windows --release --name fyclip --icon icon.png

# macOS app bundle / dmg
fyne package --os darwin --release --name fyclip --icon icon.png
```

### Manual Build

#### Linux

```bash
go build -ldflags="-s -w" -o fyclip
```

#### Windows

```bash
go build -ldflags="-s -w -H=windowsgui" -o fyclip.exe
```

#### macOS

```bash
go build -ldflags="-s -w" -o fyclip
```

### Cross-Platform Build

Use `fyne-cross` for easy cross-compilation:

```bash
# Install fyne-cross
go install github.com/fyne-io/fyne-cross@latest

# Build for Linux
fyne-cross linux -arch=amd64

# Build for Windows
fyne-cross windows -arch=amd64

# Build for macOS
fyne-cross darwin -arch=amd64
```

## Usage

### Keyboard Shortcuts

- **Ctrl+C**: Copy selected item to clipboard
- **Delete**: Delete selected item
- **Ctrl+F**: Focus search bar

### Features

1. **Pin Items**: Click the pin button to keep items at the top
2. **Search**: Type in the search bar to filter items
3. **Favorites Filter**: Click "Favorites" to show pinned items only
4. **Preview**: Select an item to see full content
5. **Export**: Click "Export" to save selected text or image
6. **Pause Monitoring**: Use "Pause 5m" to temporarily stop capturing
7. **History Limit**: Configure max unpinned history via toolbar settings
8. **Clear History**: Remove all unpinned items
9. **System Tray**: Minimize to tray, configure autostart/pause

## Configuration

Settings are automatically saved to:

- **Linux**: `~/.fyclip/`
- **Windows**: `%USERPROFILE%\.fyclip\`
- **macOS**: `~/.fyclip/`

## Architecture Highlights

### Modular Design

- **Separation of Concerns**: Each module has a single responsibility
- **Clean Interfaces**: Well-defined APIs between components
- **Testability**: Easy to unit test individual modules

### Performance Optimizations

- **Debounced Updates**: UI updates are batched (50ms debounce)
- **Async Operations**: Storage and clipboard ops don't block UI
- **Efficient Filtering**: Smart search with early returns
- **Thread-Safe**: Proper mutex usage throughout
- **Selection Fast Path**: Selecting list items avoids redundant full-window refreshes
- **Duplicate Promotion**: Existing duplicates move to latest with notification

### Thread Safety

- All shared state protected with `sync.RWMutex`
- Proper locking hierarchy to prevent deadlocks
- Channel-based communication for cross-goroutine updates

## Development

### Adding New Features

1. **New clipboard item types**: Extend `clipboard/item.go`
2. **UI components**: Add to `internal/ui/`
3. **Platform features**: Implement in `internal/platform/`

### Code Style

- Follow Go conventions
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

## Troubleshooting

### Clipboard not working on Linux

Make sure you have the required clipboard tools:

```bash
# Check for xclip
which xclip

# Check for wl-paste
which wl-paste
```

### Build errors

Make sure you have the latest Fyne dependencies:

```bash
go get -u fyne.io/fyne/v2@latest
go mod tidy
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a history of changes.

## License

MIT License - See LICENSE file for details

## Author

**Sarwar Hossain**

- Email: sarwarhridoy4@gmail.com
- GitHub: [@Sarwarhridoy4](https://github.com/Sarwarhridoy4)

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - Cross-platform GUI toolkit
- Uses [golang.design/x/clipboard](https://github.com/golang-design/clipboard) for clipboard access
