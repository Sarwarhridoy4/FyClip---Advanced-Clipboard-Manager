# FyClip - Advanced Clipboard Manager

A powerful, cross-platform clipboard manager built with Go and Fyne that automatically tracks your clipboard history, provides instant search, and persists data between sessions.

![FyClip Screenshot](screenshot.png)

## ✨ Features

### Core Functionality
- **🔄 Automatic Clipboard Monitoring**: Captures all clipboard content in real-time (500ms intervals)
- **💾 Persistent Storage**: Saves clipboard history to JSON file, restored on restart
- **🔍 Real-time Search**: Instantly filter clipboard history with live search
- **📋 One-click Copy**: Select any historical item and copy it back to clipboard
- **🗑️ Item Management**: Delete individual items or clear entire history
- **📊 Live Statistics**: Shows current item count in status bar

### Smart Features
- **🚫 Duplicate Prevention**: Automatically prevents duplicate entries, moves existing items to top
- **💡 Memory Management**: Limits history to 1000 items to prevent memory bloat
- **📏 Content Filtering**: Ignores empty clipboard and very large content (>10KB)
- **🎨 Smart Display**: Truncates long text and replaces newlines for better readability
- **⚡ Performance Optimized**: Asynchronous file operations, non-blocking UI
- **🔒 Thread-Safe**: Proper synchronization for reliable multi-threaded operation

### User Interface
- **🎯 Clean Design**: Intuitive layout with search bar, list view, and action buttons
- **🖼️ Custom Icon Support**: Load your own icon.png for personalized branding
- **📱 Responsive Layout**: Adapts to different window sizes
- **⌨️ Keyboard Friendly**: Easy navigation and selection

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or later
- Git (for cloning)

### Installation

#### Option 1: Download Pre-built Binaries
Visit the [Releases](https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/releases) page and download the appropriate binary for your platform.

#### Option 2: Build from Source
```bash
# Clone the repository
git clone https://github.com/yourusername/fyclip.git
cd fyclip

# Download dependencies
go mod tidy

# Run directly
go run main.go

# Or build executable
go build -o fyclip main.go
```

## 🔧 Build Instructions

### For All Platforms

#### Windows
```bash
# Build for Windows (from any platform)
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o fyclip.exe main.go

# Build with icon (requires rsrc tool)
# Install rsrc: go install github.com/akavel/rsrc@latest
# rsrc -manifest fyclip.manifest -ico icon.ico -o rsrc.syso
# go build -ldflags="-H windowsgui" -o fyclip.exe main.go
```

#### macOS
```bash
# Build for macOS (64-bit Intel)
GOOS=darwin GOARCH=amd64 go build -o fyclip-mac-intel main.go

# Build for macOS (Apple Silicon/M1/M2)
GOOS=darwin GOARCH=arm64 go build -o fyclip-mac-apple main.go

# Create macOS app bundle
mkdir -p FyClip.app/Contents/MacOS
mkdir -p FyClip.app/Contents/Resources
cp fyclip-mac-* FyClip.app/Contents/MacOS/fyclip
cp icon.png FyClip.app/Contents/Resources/
```

#### Linux
```bash
# Build for Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o fyclip-linux main.go

# Build for Linux (32-bit)
GOOS=linux GOARCH=386 go build -o fyclip-linux-32 main.go

# Build for Linux ARM (Raspberry Pi)
GOOS=linux GOARCH=arm GOARM=7 go build -o fyclip-linux-arm main.go

# Build for Linux ARM64
GOOS=linux GOARCH=arm64 go build -o fyclip-linux-arm64 main.go
```

#### FreeBSD
```bash
# Build for FreeBSD
GOOS=freebsd GOARCH=amd64 go build -o fyclip-freebsd main.go
```

### Cross-Compilation Script
Create a `build.sh` script for building all platforms at once:

```bash
#!/bin/bash

# Create build directory
mkdir -p builds

# Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o builds/fyclip-windows.exe main.go

# macOS
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o builds/fyclip-macos-intel main.go

echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o builds/fyclip-macos-apple main.go

# Linux
echo "Building for Linux (64-bit)..."
GOOS=linux GOARCH=amd64 go build -o builds/fyclip-linux-amd64 main.go

echo "Building for Linux (32-bit)..."
GOOS=linux GOARCH=386 go build -o builds/fyclip-linux-386 main.go

echo "Building for Linux ARM..."
GOOS=linux GOARCH=arm GOARM=7 go build -o builds/fyclip-linux-arm main.go

echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o builds/fyclip-linux-arm64 main.go

# FreeBSD
echo "Building for FreeBSD..."
GOOS=freebsd GOARCH=amd64 go build -o builds/fyclip-freebsd main.go

echo "Build complete! Check the builds/ directory."
```

## 📁 File Structure
```
fyclip/
├── main.go              # Main application code
├── icon.png             # Application icon (optional)
├── go.mod              # Go module dependencies
├── go.sum              # Dependency checksums
├── README.md           # This file
├── screenshot.png      # Application screenshot
└── builds/             # Built binaries (after running build script)
    ├── fyclip-windows.exe
    ├── fyclip-macos-intel
    ├── fyclip-macos-apple
    ├── fyclip-linux-amd64
    └── ...
```

## 🎯 Usage

1. **Launch the Application**: Run the executable or use `go run main.go`

2. **Automatic Monitoring**: Start copying text - it will automatically appear in the history

3. **Search**: Use the search bar at the top to filter clipboard items

4. **Copy Items**: Select any item from the list and click "Copy Selected"

5. **Manage Items**: 
   - Delete individual items with "Delete Selected"
   - Clear all history with "Clear All"

6. **View Statistics**: Check the status bar for current item count

## 📊 Data Storage

- **Location**: `~/clipboard_history.json` (user's home directory)
- **Format**: JSON array of strings
- **Persistence**: History is automatically saved and restored between sessions
- **Backup**: You can manually backup the JSON file to preserve history

## ⚙️ Configuration

### Custom Icon
Place an `icon.png` file in the same directory as the executable to use a custom icon.

### Memory Limits
- Maximum clipboard content size: 10KB per item
- Maximum history items: 1000 entries
- Monitoring interval: 500ms

## 🔧 Development

### Dependencies
```bash
# Core dependencies
go get fyne.io/fyne/v2/app
go get fyne.io/fyne/v2/widget
go get fyne.io/fyne/v2/container
```

### Running Tests
```bash
go test ./...
```

### Code Structure
- **Thread-safe design**: Uses mutexes for concurrent access
- **Clean separation**: UI components separated from business logic  
- **Error handling**: Comprehensive error handling throughout
- **Resource management**: Proper cleanup and shutdown handling

## 🐛 Troubleshooting

### Common Issues

**Application won't start**
- Ensure you have the required Go version (1.19+)
- Check that all dependencies are installed: `go mod tidy`

**Clipboard not monitored**
- Some Linux distributions require additional permissions
- Try running with elevated privileges if needed

**Icon not showing**
- Ensure `icon.png` is in the same directory as the executable
- Check that the PNG file is valid and not corrupted

**Performance issues**
- Clear clipboard history if it becomes too large
- Check available system memory

### Platform-Specific Notes

**Windows**
- Windows Defender might flag the executable initially
- Use `-ldflags="-H windowsgui"` to hide console window

**macOS**
- May require approval for accessibility permissions
- Create proper app bundle for distribution

**Linux**
- Some desktop environments may need additional clipboard access
- Install required system libraries: `apt-get install libgl1-mesa-dev xorg-dev`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit your changes: `git commit -am 'Add feature'`
4. Push to the branch: `git push origin feature-name`
5. Submit a pull request

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/fyclip/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/fyclip/discussions)
- **Email**: your.email@domain.com

## 🎉 Acknowledgments

- Built with [Fyne](https://fyne.io/) - Cross-platform GUI framework
- Inspired by various clipboard managers across different platforms
- Thanks to the Go and Fyne communities for excellent documentation and support

---

**FyClip** - Making clipboard management effortless across all platforms! 🚀
