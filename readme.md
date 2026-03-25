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

> A modular, high-performance clipboard manager built with Go and Fyne v2.7+

**Current Version**: 2.2.0

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

FyClip is a feature-rich clipboard manager designed for power users. Here's a comprehensive breakdown of what it offers:

### Core Functionality

| Feature | Description |
|---------|-------------|
| 📋 **Clipboard History** | Automatically saves text, images, HTML, and files |
| 🔍 **Enhanced Search** | Regex, case-sensitive, and fuzzy matching |
| 📌 **Pin Items** | Keep important items at the top |
| ⭐ **Favorites View** | Toggle pinned-only view instantly |
| ❌ **Clear Search** | One-click reset for the search box |

### Content Support

| Feature | Description |
|---------|-------------|
| 🖼️ **Image Support** | Preview and save clipboard images |
| 📝 **HTML Support** | Capture and preserve HTML formatting |
| 📁 **File History** | Track files copied from file manager |
| 📝 **Markdown Preview** | Markdown content renders correctly in preview pane |
| 📤 **Unified Export** | Export selected text or images from one action |

### Organization & Management

| Feature | Description |
|---------|-------------|
| 🏷️ **Smart Categories** | Auto-categorize content (Links, Code, Contacts, etc.) |
| 🏷️ **Custom Tags** | Add custom tags to organize clipboard items |
| 📝 **Snippets** | Save and expand text templates with variables |
| 📦 **Bulk Operations** | Multi-select items for batch delete/pin/unpin |

### Security

| Feature | Description |
|---------|-------------|
| 🔒 **Encrypted Storage** | AES-256-GCM encryption at rest |
| ☁️ **Encrypted Backup** | Password-protected backup and restore |
| 🛡️ **Sensitive Data Detection** | Auto-detect credit cards, SSN, API keys |

### System Integration

| Feature | Description |
|---------|-------------|
| 🚀 **AutoStart** | Launch on system startup |
| ⏸️ **Pause Capture** | Pause monitoring for 5 minutes from toolbar/tray |
| 🖥️ **System Tray** | Recent items submenu, Clear History action |
| ⬆️ **Auto Update** | Check for and install updates from GitHub releases |

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
| 🔒 **Thread-Safe** | Proper concurrency handling |
| 🖼️ **Image Thumbnails** | 150x150px thumbnails for efficient list display |
| 📦 **Compression** | Gzip compression before encryption for smaller storage |
| 💾 **Memory Optimized** | Lazy loading and efficient data structures |

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
| **.deb** | `sudo dpkg -i fyclip_<version>_<arch>.deb` |
| **.AppImage** | `chmod +x fyclip_<version>_<arch>.AppImage && ./fyclip_<version>_<arch>.AppImage` |

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
| Fyne | 2.5+ | UI framework |

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
./build.sh 2.2.0
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
| **Efficient Filtering** | Search avoids repeated lowercasing and minimizes allocation churn |
| **Object Pool** | `sync.Pool` for Item reuse to reduce GC pressure |
| **Regex Cache** | Compiled regex patterns cached for faster repeated searches |
| **Fuzzy Search** | Optimized subsequence matching with reduced allocations |
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
| **AES-256-GCM Encryption** | All data encrypted at rest |
| **Sensitive Data Detection** | Auto-detect credit cards, SSN, API keys |
| **Secure Wipe** | Clear sensitive data from memory |
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
| Fyne | 2.5+ | UI framework |

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
