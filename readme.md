# FyClip - Advanced Clipboard Manager

<p align="center">
  <img src="internal/app/assets/icon.png" alt="FyClip Logo" width="128" height="128"/>
  <br>
  <a href="https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/releases/latest">
    <img src="https://img.shields.io/github/v/release/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager?include_prereleases&style=flat" alt="GitHub release">
  </a>
  <a href="https://goreportcard.com/report/github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager">
    <img src="https://goreportcard.com/badge/github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager" alt="Go Report Card">
  </a>
  <a href="https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager" alt="License">
  </a>
  <a href="https://discord.gg/fyclip">
    <img src="https://img.shields.io/discord/123456789" alt="Discord">
  </a>
</p>

> A secure, fast clipboard manager for text, images, HTML, and files, built with Go and Fyne v2.7+

**Current Version**: 2.2.2

---

## Table of Contents

- [Features](#features)
- [Screenshots](#screenshots)
- [Quick Start](#quick-start)
- [Installation](#installation)
  - [Linux](#linux)
  - [Windows](#windows)
  - [macOS](#macos)
  - [Building from Source](#building-from-source)
- [Building for Release](#building-for-release)
- [Usage](#usage)
  - [Keyboard Shortcuts](#keyboard-shortcuts)
  - [Bulk Operations](#bulk-operations)
  - [Snippets](#snippets)
- [Configuration](#configuration)
- [Architecture](#architecture)
  - [Project Structure](#project-structure)
  - [Design Principles](#design-principles)
  - [Performance Optimizations](#performance-optimizations)
  - [Security](#security)
- [Development](#development)
  - [Prerequisites](#prerequisites)
  - [Makefile Targets](#makefile-targets)
  - [Adding New Features](#adding-new-features)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [Changelog](#changelog)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## Features

FyClip is built for people who copy a lot and want fast recall, reliable history, and safer local storage.

### Highlights

- **Rich clipboard history** for text, images, HTML, and files
- **Fast search** with regex, fuzzy matching, and case-sensitive modes
- **Pinned items, tags, categories, and snippets** for organization
- **Encrypted local storage and backups** with safer clipboard write paths
- **Cross-platform desktop integration** with tray controls, pause capture, and auto-update support

### Feature Overview

| Area | Included |
|------|----------|
| **Capture** | Text, images, HTML, and file history |
| **Search** | Regex, fuzzy matching, case-sensitive filtering, clear search |
| **Organize** | Pinning, pinned-only view, smart categories, custom tags, snippets |
| **Actions** | Copy, export, bulk select, bulk delete, pin/unpin |
| **Security** | AES-256-GCM storage, PBKDF2 key derivation, encrypted backups, sensitive-pattern detection, clipboard/path validation |
| **System** | Autostart, pause capture, system tray actions, GitHub-based auto updates |

### Auto Update Feature

The auto-update feature allows you to check for and install updates directly from GitHub releases.

#### How It Works

1. **Version Detection**: Reads the current version from embedded `internal/version/version.go` file (generated during build)
2. **GitHub Check**: Fetches the latest release from GitHub API
3. **Version Comparison**: Compares semantic versions, including `v` prefixes and pre-release handling
4. **Request Optimization**: Uses in-memory caching and rate limiting to reduce repeated GitHub API traffic
5. **Asset Selection**: Scores release assets for the active OS and architecture, preferring native installer formats like `.deb`, `.exe`, and `.dmg`
6. **Update Available**: Prompts when a newer compatible release is found
7. **Up to Date**: Reports when the current version is already the latest supported release

#### Why it is efficient

- Caches successful update checks in memory
- Applies rate limiting to repeated GitHub API calls
- Scores release assets concurrently for the active OS and architecture
- Prefers native installer formats when multiple assets are available

#### Usage

**From UI:**
- Click "Help" → "Check for Updates" in the menu

**From Terminal:**
```bash
# Check for updates
fyclip --check-update

# Download and install updates
fyclip --update
```

#### Supported Platforms

| Platform | Package Formats |
|----------|----------------|
| Linux | Snap, .deb, .AppImage |
| Windows | .exe, .msi |
| macOS | .dmg |

### User Interface

| Feature | Description |
|---------|-------------|
| 🎨 **Theme Support** | Light, Dark, and System theme modes |
| 🎨 **Modern UI** | Dark theme with responsive design |
| ⌨️ **Keyboard Navigation** | Arrow keys, Enter, Delete, Escape, Space, Home/End, F1 |
| 🕒 **Relative Time** | List rows show recency and copy frequency |
| 💾 **Persistent Storage** | History saved across sessions |

### Performance

| Feature | Description |
|---------|-------------|
| ⚡ **Debounced Updates** | Batched UI updates for smooth performance |
| ⚡ **Async Operations** | Non-blocking clipboard monitoring |
| ⚡ **O(1) Lookups** | Hash map-based duplicate detection |
| 🧠 **Adaptive Memory Pressure Handling** | Multi-level cleanup responds to sustained memory growth |
| 🕒 **LRU History Retention** | Frequently used items are preserved longer than stale history |
| 🔒 **Thread-Safe** | Proper concurrency handling |
| 🖼️ **Image Thumbnails** | 150x150px thumbnails for efficient list display |
| 📦 **Compression** | Gzip compression before encryption for smaller storage |
| 💾 **Memory Optimized** | Differential indexing, object reuse, and reduced clipboard write copies |

### Security At A Glance

| Feature | Description |
|---------|-------------|
| 🔒 **Encrypted Storage** | Clipboard history is encrypted at rest |
| 🔐 **PBKDF2 Key Derivation** | Derived keys with migration support for legacy installs |
| ☁️ **Encrypted Backup** | Password-protected backup and restore |
| 🧼 **Memory-Safer Writes** | Programmatic copy hashing and temporary buffer wiping where practical |
| 🚫 **Validation Guards** | Clipboard size limits plus file-path and command validation |

---

## Screenshots

<div align="center">
  <img src="internal/app/assets/screenshots/screenshot1.png" alt="Main Window - Clipboard History with Search and Preview" width="800"/>
  <p><em>Main Window - Clipboard History with Search and Preview</em></p>
</div>

<div align="center">
  <img src="internal/app/assets/screenshots/screenshot2.png" alt="Preview Pane with JSON Formatting" width="800"/>
  <p><em>Preview Pane with JSON Formatting</em></p>
</div>

<div align="center">
  <img src="internal/app/assets/screenshots/screenshot3.png" alt="Quick Panel - Global Hotkey Access" width="800"/>
  <p><em>Quick Panel - Global Hotkey Access</em></p>
</div>

---

## Quick Start

### Linux

```bash
# Install dependencies (X11)
sudo apt install xclip

# Or for Wayland
sudo apt install wl-clipboard

# Run the application
./fyclip
```

### Windows

```bash
# Simply run the executable
fyclip.exe
```

### macOS

```bash
# Run the application
./fyclip
```

---

## Installation

### Pre-built Packages

#### Linux

| Format | Command |
|--------|---------|
| **Snap** | `sudo snap install fyclip` |
| **PPA** | `sudo add-apt-repository ppa:sarwar-hossain/fyclip && sudo apt update && sudo apt install fyclip` |
| **.deb** | `sudo dpkg -i fyclip_<version>_<arch>.deb` |
| **.AppImage** | `chmod +x fyclip_<version>_<arch>.AppImage && ./fyclip_<version>_<arch>.AppImage` |

**PPA Repository**: https://launchpad.net/~sarwar-hossain/+archive/ubuntu/fyclip

Download from [Releases](https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/releases)

#### Windows

Download the installer from [Releases](https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/releases)

#### macOS

```bash
# Using Homebrew (if available)
brew install fyclip
```

### Building from Source

#### Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | Programming language |
| Fyne | 2.7+ | UI framework |

#### Linux Dependencies

**Ubuntu/Debian:**
```bash
# For X11
sudo apt install xclip

# For Wayland
sudo apt install wl-clipboard
```

**Arch Linux:**
```bash
# For X11
sudo pacman -S xclip

# For Wayland
sudo pacman -S wl-clipboard
```

**Fedora:**
```bash
# For X11
sudo dnf install xclip

# For Wayland
sudo dnf install wl-clipboard
```

#### Build Steps

```bash
# 1. Clone the repository
git clone https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager.git
cd FyClip---Advanced-Clipboard-Manager

# 2. Install dependencies
go mod download

# 3. Build
make build

# 4. Run
./fyclip
```

---

## Building for Release

### Using Build Script (Recommended for Linux)

The build script follows Fyne's official Linux packaging flow:

```bash
# Build with default version
./build.sh

# Build with explicit version
./build.sh 2.2.2
```

This produces:
- `dist/fyclip_<version>_<arch>.deb` - Debian package
- `dist/fyclip_<version>_<arch>.AppImage` - AppImage

**Requirements:**
- `go`
- `fyne` CLI (`go install github.com/fyne-io/fyne@latest`)
- `dpkg-deb`
- `appimagetool`
- `tar`

### Using Makefile

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Package for distribution
make package

# Create release
make release
```

### Using Fyne Native Packaging

```bash
# Linux tar package
fyne package --os linux --release --name fyclip --icon icon.png

# Windows installer
fyne package --os windows --release --name fyclip --icon icon.png

# macOS app bundle
fyne package --os darwin --release --name fyclip --icon icon.png
```

### Cross-Platform Build with fyne-cross

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

---

## Usage

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `↑/↓` | Navigate through clipboard items |
| `Enter` | Copy selected item to clipboard |
| `Delete` | Delete selected item |
| `Space` | Pin/unpin selected item |
| `Escape` | Clear search / Close panel |
| `Home` | Go to first item |
| `End` | Go to last item |
| `F1` | Focus search bar |
| `Ctrl+F` | Focus search bar |
| `Ctrl+Shift+V` | Open quick panel |

### Bulk Operations

| Shortcut | Action |
|----------|--------|
| `Ctrl+Click` | Add/remove item from selection |
| Toolbar **Select** button | Enter selection mode |

### Features Guide

1. **Pin Items**: Click the pin button or press Space to keep items at the top
2. **Search**: Type in the search bar to filter items (supports regex, case-sensitive, fuzzy)
3. **Favorites Filter**: Click "Favorites" to show pinned items only
4. **Preview**: Select an item to see full content (JSON pretty-printed automatically)
5. **Export**: Click "Export" to save selected text or image
6. **Pause Monitoring**: Use "Pause 5m" to temporarily stop capturing
7. **History Limit**: Configure max unpinned history via toolbar settings
8. **Clear History**: Remove all unpinned items
9. **System Tray**: Minimize to tray, configure autostart/pause, access recent items
10. **Snippets**: Create and manage text templates
11. **Backup**: Create encrypted backups of your history
12. **Categories**: Auto-categorized content (Links, Code, Contacts, Images, Files, Text)
13. **Tags**: Add custom tags to organize items
14. **Theme**: Switch between Light, Dark, and System themes
15. **Bulk Operations**: Multi-select items for batch actions

### Snippets

Snippets allow you to create reusable text templates with dynamic variables.

#### Creating a Snippet

1. Click the "Snippets" button in the toolbar
2. Click "Add Snippet" button
3. Fill in the details:
   - **Title**: A descriptive name (e.g., "Email Signature")
   - **Content**: The template text with optional variables
   - **Abbreviation** (optional): A short trigger word for quick access
   - **Category** (optional): Group snippets by category

#### Available Variables

| Variable | Description | Example Output |
|----------|-------------|----------------|
| `{{date}}` | Current date | 2026-03-25 |
| `{{time}}` | Current time | 14:30:45 |
| `{{datetime}}` | Full date and time | 2026-03-25 14:30:45 |
| `{{year}}` | Current year | 2026 |
| `{{month}}` | Current month (01-12) | 03 |
| `{{day}}` | Current day (01-31) | 25 |
| `{{clipboard}}` | Current clipboard content | (varies) |

#### Example Snippet

```
Title: Email Signature
Abbreviation: sig
Category: General
Content:
Best regards,
{{name}}
{{date}}
```

When used, this expands to:
```
Best regards,
John Doe
2026-03-25
```

---

## Configuration

Settings are automatically saved to:

| Platform | Path |
|----------|------|
| Linux | `~/.fyclip/` |
| Windows | `%USERPROFILE%\.fyclip\` |
| macOS | `~/.fyclip/` |

---

## Architecture

### Project Structure

```
fyclip/
├── main.go                      # Application entry point
├── Makefile                     # Build automation
├── go.mod                       # Go module dependencies
├── icon.png                     # Application icon
├── FyneApp.toml                 # Fyne application configuration
├── build.sh                     # Release build script
│
├── internal/
│   ├── app/
│   │   ├── app.go              # App initialization & lifecycle
│   │   └── single_instance.go  # Single instance enforcement
│   │
│   ├── clipboard/
│   │   ├── item.go             # Clipboard item types (Text, Image, HTML, File)
│   │   ├── manager.go          # Core manager logic & state
│   │   ├── monitor.go          # Clipboard monitoring & change detection
│   │   ├── native.go           # Platform-specific clipboard operations
│   │   ├── storage.go          # Persistence layer with encryption
│   │   ├── snippet.go          # Snippet management & expansion
│   │   ├── exclusion.go        # Pattern exclusion rules
│   │   ├── search.go           # Enhanced search (regex, fuzzy)
│   │   ├── backup.go           # Encrypted backup & restore
│   │   ├── sensitive.go        # Sensitive data detection
│   │   ├── pool.go             # Object pooling for performance
│   │   └── validation.go       # Input validation
│   │
│   ├── config/
│   │   ├── config.go           # Configuration management
│   │   └── validation.go       # Config validation
│   │
│   ├── errors/
│   │   └── errors.go           # Custom error types
│   │
│   ├── logger/
│   │   └── logger.go           # Structured logging with file rotation
│   │
│   ├── metrics/
│   │   └── metrics.go          # Application metrics
│   │
│   ├── ui/
│   │   ├── window.go           # Main window management
│   │   ├── list.go             # History list widget
│   │   ├── preview.go          # Preview pane with formatting
│   │   ├── toolbar.go          # Action buttons & controls
│   │   ├── search.go           # Search bar component
│   │   ├── status.go           # Status bar
│   │   ├── dialogs.go          # Dialogs & utilities
│   │   ├── quickpanel.go       # Quick access panel
│   │   └── update_dialogs.go   # Update notification dialogs
│   │
│   ├── platform/
│   │   ├── autostart.go        # Autostart interface
│   │   ├── autostart_linux.go  # Linux implementation
│   │   ├── autostart_windows.go # Windows implementation
│   │   ├── autostart_darwin.go # macOS implementation
│   │   └── utils.go            # Platform utilities
│   │
│   ├── tray/
│   │   └── tray.go             # System tray integration
│   │
│   ├── update/
│   │   └── checker.go          # Auto-update checking
│   │
│   └── testutil/
│       └── testutil.go         # Testing utilities
│
└── README.md
```

### Design Principles

#### Modular Architecture
- **Separation of Concerns**: Each module has a single responsibility
- **Clean Interfaces**: Well-defined APIs between components
- **Testability**: Easy to unit test individual modules

#### Code Organization
- Internal packages (`internal/`) for implementation details
- Clear dependency direction: `main.go` → `app` → `clipboard`/`ui` → `config`/`logger`
- Feature flags for optional functionality

### Performance Optimizations

| Optimization | Description |
|--------------|-------------|
| **Debounced Updates** | UI updates are batched (50ms debounce) |
| **Coalesced Saves** | History persistence requests are serialized and debounced (250ms) |
| **O(1) Lookups** | Hash maps for duplicate detection and item access |
| **Differential Index Rebuilds** | Modified indices are refreshed selectively instead of full rebuilds on every change |
| **Efficient Filtering** | Search avoids repeated lowercasing and minimizes allocation churn |
| **Object Pool** | `sync.Pool` for Item reuse to reduce GC pressure |
| **Regex Cache** | Compiled regex patterns cached for faster repeated searches |
| **Fuzzy Search** | Optimized subsequence matching with reduced allocations |
| **Adaptive Memory Pressure Handling** | Cleanup becomes more aggressive as allocation pressure rises |
| **LRU Trimming** | History eviction preserves frequently accessed items over merely recent ones |
| **Update Check Caching** | Repeated GitHub release checks are cached and rate-limited |
| **Concurrent Asset Scoring** | Release assets are processed concurrently during platform-specific update selection |
| **Thread-Safe** | Proper mutex usage throughout |
| **Selection Fast Path** | Selecting list items avoids redundant full-window refreshes |
| **Duplicate Promotion** | Existing duplicates move to latest with notification |

#### Benchmark Results

Run the benchmarks to see performance improvements:

```bash
go test -bench 'Benchmark(UpdateFilteredSearch1000|AddItemWithDuplicateScan1000|StorageSave1000)
 -benchmem ./internal/clipboard
```

Latest measured deltas:
- `BenchmarkUpdateFilteredSearch1000`: `604006 ns/op` -> `37772 ns/op` (~16x faster)
- `BenchmarkUpdateFilteredSearch1000` allocations: `1020 allocs/op` -> `0 allocs/op`
- `BenchmarkAddItemWithDuplicateScan1000`: `966.1 ns/op` -> `1523 ns/op` (small regression, low absolute cost)
- `BenchmarkStorageSave1000`: raw `Storage.Save` micro-benchmark unchanged/slower, but save requests are now coalesced in runtime manager flow

### Security

| Feature | Description |
|---------|-------------|
| **AES-256-GCM Encryption** | Clipboard history is encrypted at rest |
| **PBKDF2 Key Derivation** | Storage keys are derived with migration support for legacy key files |
| **Sensitive Data Detection** | Auto-detect credit cards, SSN, API keys |
| **Secure Wipe** | Temporary sensitive buffers are cleared after use where practical |
| **Clipboard Size Validation** | Large text, image, and path payloads are rejected before clipboard writes |
| **Programmatic Copy Hashing** | Monitor state tracks programmatic writes without storing raw clipboard content |
| **Path Validation** | File-oriented clipboard and open-location operations validate paths defensively |
| **Command Validation** | Platform-specific command arguments are sanitized and checked before execution |
| **Password-Protected Backups** | Optional encryption for backups |

### Thread Safety

- All shared state protected with `sync.RWMutex`
- Proper locking hierarchy to prevent deadlocks
- Channel-based communication for cross-goroutine updates

---

## Development

### Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | Programming language |
| Fyne | 2.7+ | UI framework |

### Makefile Targets

```bash
make help    # Show available targets
make build   # Build the application
make test    # Run tests
make lint    # Run linter
make clean   # Clean build artifacts
```

### Adding New Features

1. **New clipboard item types**: Extend `internal/clipboard/item.go`
2. **UI components**: Add to `internal/ui/`
3. **Platform features**: Implement in `internal/platform/`

### Code Style

- Follow Go conventions
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

---

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

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a history of changes.

---

## License

MIT License - See [LICENSE](Licence) file for details

---

## Acknowledgments

- [Fyne](https://fyne.io/) - Cross-platform GUI library
- [Go](https://golang.org/) - Programming language
- All contributors and testers

---

<p align="center">
  <strong>Made with ❤️ by <a href="https://github.com/Sarwarhridoy4">Sarwar Hossain</a></strong>
</p>
