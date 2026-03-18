# FyClip - Scope of Improvement

This document outlines potential areas for improvement and feature enhancements for the FyClip Advanced Clipboard Manager project. The suggestions are categorized by priority and impact.

---

## 1. High Priority Improvements

### 1.1 Unit Testing Framework

**Current State**: The project has benchmark tests (`perf_bench_test.go`) but lacks unit tests.

**Suggested Improvements**:
- Add unit tests for core modules:
  - [`clipboard/item.go`](internal/clipboard/item.go) - Test `DisplayText()`, `Size()`, `TimeAgo()` methods
  - [`clipboard/manager.go`](internal/clipboard/manager.go) - Test `AddItem()`, `Delete()`, `TogglePin()`, `SetSearch()` methods
  - [`clipboard/storage.go`](internal/clipboard/storage.go) - Test `Load()`, `Save()`, `encrypt()`, `decrypt()` functions
- Use table-driven tests for edge cases
- Add integration tests for manager-storage interaction
- Implement test coverage monitoring via CI/CD

### 1.2 Hash/Index Maps for O(1) Operations

**Current State**: Duplicate detection and pin/delete operations use linear O(n) scans ([`manager.go:184-190`](internal/clipboard/manager.go:184), [`manager.go:419-425`](internal/clipboard/manager.go:419))

**Suggested Improvements**:
```go
// Add to Manager struct
hashIndexMap map[string]int  // hash -> index in history
idIndexMap    map[string]int  // id -> index in history
```

- Implement hash-based O(1) duplicate lookups
- Implement ID-based O(1) item retrieval for pin/toggle operations
- Maintain maps during add/delete operations
- Update benchmarks to measure improvement

### 1.3 Comprehensive Configuration System

**Current State**: Only history limit is configurable through the UI ([`toolbar.go:206-229`](internal/ui/toolbar.go:206))

**Suggested Improvements**:
- Create `internal/config/config.go` for persistent settings:
  ```go
  type Config struct {
      MaxHistoryItems    int           `json:"max_history_items"`
      MonitoringInterval time.Duration `json:"monitoring_interval"`
      StartMinimized     bool          `json:"start_minimized"`
      Theme              string        `json:"theme"` // "dark", "light", "system"
      PauseDuration      time.Duration `json:"pause_duration"`
      AutoSaveInterval   time.Duration `json:"auto_save_interval"`
      EnableNotifications bool         `json:"enable_notifications"`
  }
  ```
- Add settings dialog with all configurable options
- Persist configuration to `~/.fyclip/config.json`
- Add configuration migration support for future versions

---

## 2. Medium Priority Improvements

### 2.1 Global Hotkey Quick Panel

**Current State**: Mentioned in README as planned enhancement.

**Suggested Improvements**:
- Implement a global hotkey (e.g., `Ctrl+Shift+V`) that shows a quick-access popup
- Display recent clipboard items in an overlay panel
- Allow quick paste by pressing number keys (1-9)
- Use platform-specific hotkey registration:
  - Linux: `github.com/mholt/archiver/v3` or `github.com/rivo/tview` combined with global hotkey package
  - Windows: `github.com/getlantern/showkey` or `github.com/go-void/globalkey`
  - macOS: `github.com/eafer/ApplePressOrKey`

### 2.2 Snippets/Templates System

**Current State**: Not implemented.

**Suggested Improvements**:
- Add snippet management UI in toolbar or separate dialog
- Support categories/folders for organizing snippets
- Implement template variables:
  - `{{date}}` - Current date
  - `{{time}}` - Current time
  - `{{clipboard}}` - Current clipboard content
  - `{{cursor}}` - Cursor position after paste
- Add keyboard shortcut to expand snippets (e.g., type abbreviation and press Tab)

### 2.3 Pattern Exclusion Rules

**Current State**: Not implemented.

**Suggested Improvements**:
- Add exclusion rules configuration:
  - Regex-based content filtering
  - App-specific exclusions (e.g., don't capture from password managers)
  - Size-based filtering (ignore items larger than X MB)
- Add UI for managing exclusion patterns
- Implement "sensitive mode" that pauses capture temporarily:
  ```go
  // Auto-pause when specific apps are focused
  type ExclusionRule struct {
      Pattern   string `json:"pattern"`
      Type      string `json:"type"` // "regex", "app", "size"
      Value     string `json:"value"`
      Enabled   bool   `json:"enabled"`
  }
  ```

### 2.4 Enhanced Search Functionality

**Current State**: Basic substring search implemented ([`manager.go:231-240`](internal/clipboard/manager.go:231))

**Suggested Improvements**:
- Add regex search mode toggle
- Add case-sensitive search option
- Implement fuzzy matching for typo tolerance
- Add search history (recent searches dropdown)
- Add tag-based filtering for items
- Implement full-text search index using:
  - Simple in-memory index for text items
  - Tokenize and index content for faster searches

---

## 3. Feature Enhancements

### 2.5 Encrypted Import/Export Backup

**Current State**: Export functionality exists for individual items, but no encrypted backup system.

**Suggested Improvements**:
- Implement bulk export with encryption:
  ```go
  type Backup struct {
      Version     string    `json:"version"`
      Timestamp   time.Time `json:"timestamp"`
      Items       []Item    `json:"items"`
      Config      Config    `json:"config"`
      Checksum    string    `json:"checksum"`
  }
  ```
- Add password-protected backup option using user's choice of:
  - AES-256-GCM with password-derived key (PBKDF2)
  - Or use existing encryption key from storage
- Add import functionality with merge/replace options
- Add scheduled automatic backups

### 2.6 Rich Text/HTML Support

**Current State**: Only text and images are supported.

**Suggested Improvements**:
- Add support for HTML clipboard content
- Store both plain text and HTML versions of items
- Add preview toggle between plain text and rendered HTML
- Add "Copy as Plain Text" vs "Copy as HTML" options

### 2.7 File Path History

**Current State**: Not implemented.

**Suggested Improvements**:
- Capture file paths when files are copied
- Store file metadata (name, size, path, modification time)
- Add "Open file location" action
- Add "Copy as path" vs "Copy content" options for file items

---

## 4. Code Quality Improvements

### 4.1 Error Handling Enhancement

**Current State**: Basic error handling with logging.

**Suggested Improvements**:
- Create custom error types:
  ```go
  type ClipboardError struct {
      Code    string
      Message string
      Err     error
  }
  ```
- Implement error wrapping with `fmt.Errorf("%w", ...)` pattern consistently
- Add error recovery mechanisms in goroutines
- Create error log with categories for debugging

### 4.2 Structured Logging

**Current State**: Uses basic `log` package.

**Suggested Improvements**:
- Implement structured logging with levels (debug, info, warn, error)
- Use `log/slog` (Go 1.21+) or `github.com/slog/slog`
- Add contextual fields to logs:
  ```go
  slog.Info("item added", 
      "id", item.ID, 
      "type", item.Type,
      "size", item.Size())
  ```
- Add log rotation for persistent logging
- Add log file location in configuration

### 4.3 Graceful Shutdown

**Current State**: Basic shutdown implementation ([`manager.go:658-674`](internal/clipboard/manager.go:658))

**Suggested Improvements**:
- Add context.Context for cancellation propagation
- Implement graceful goroutine termination with timeouts
- Add shutdown hooks for cleanup:
  ```go
  type ShutdownHook func() error
  func (m *Manager) AddShutdownHook(fn ShutdownHook)
  ```
- Save state verification before exit

---

## 5. UI/UX Improvements

### 5.1 Enhanced Toolbar and Actions

**Current State**: Basic toolbar with copy, pin, delete, clear, settings, export.

**Suggested Improvements**:
- Add keyboard shortcut hints on buttons
- Add right-click context menu on list items:
  - Copy, Pin/Unpin, Delete, Export, View Details
- Add bulk selection mode (multi-select with Ctrl/Shift)
- Add drag-and-drop reordering for pinned items
- Add item categories/tags management

### 5.2 Preview Pane Enhancements

**Current State**: Basic preview implemented.

**Suggested Improvements**:
- Add syntax highlighting for code content
- Add image zoom controls
- Add "Copy Selection" in preview for partial text
- Add JSON pretty-printing for JSON content
- Add hex view for binary data

### 5.3 Settings Dialog

**Current State**: Only history limit setting.

**Suggested Improvements**:
- Create comprehensive settings dialog with tabs:
  - General (history limit, startup behavior)
  - Monitoring (interval, exclusions)
  - Appearance (theme, font size)
  - Shortcuts (keyboard shortcuts configuration)
  - Storage (encryption, backup)
  - About (version, licenses)

### 5.4 System Tray Improvements

**Current State**: Basic tray menu implemented.

**Suggested Improvements**:
- Add recent items submenu in tray
- Add quick actions in tray (pause/resume, clear)
- Add tray icon badge for notification count
- Add double-click to show window

---

## 6. Performance Optimizations

### 6.1 Lazy Loading for Images

**Current State**: Images stored as base64 in memory.

**Suggested Improvements**:
- Load image data on-demand for preview
- Generate thumbnails for list display
- Use image caching with LRU eviction:
  ```go
  type ImageCache struct {
      cache *lru.Cache
      maxMB int
  }
  ```

### 6.2 Virtualized List Rendering

**Current State**: All items rendered in list.

**Suggested Improvements**:
- Implement virtual scrolling for large history lists
- Only render visible items + buffer
- Use Fyne's lazy container or custom implementation

### 6.3 Memory Optimization

**Current State**: Full history kept in memory.

**Suggested Improvements**:
- Implement memory-mapped file for large history
- Add memory usage indicator in status bar
- Add automatic memory cleanup triggers
- Compress old items in storage

---

## 7. Platform-Specific Improvements

### 7.1 Linux Enhancements

**Current State**: X11 and Wayland support with xclip/wl-clipboard fallback.

**Suggested Improvements**:
- Add native Wayland clipboard support via `wlroots`
- Detect and warn about incompatible clipboard managers
- Add support for more image formats (WebP, AVIF)
- Improve Wayland polling efficiency

### 7.2 Windows Enhancements

**Current State**: Basic Windows support.

**Suggested Improvements**:
- Add Windows notification integration (Windows Toast)
- Add Windows Jump List integration
- Add Windows accent color support for theming

### 7.3 macOS Enhancements

**Current State**: Basic macOS support.

**Suggested Improvements**:
- Add native macOS notification center
- Add Touch Bar support
- Add Spotlight integration

---

## 8. Security Improvements

### 8.1 Sensitive Data Handling

**Current State**: All clipboard data encrypted at rest.

**Suggested Improvements**:
- Add "sensitive" flag for items (never save certain content)
- Add password manager detection (don't capture from 1Password, Bitwarden, etc.)
- Add secure memory handling for sensitive data:
  ```go
  func (i *Item) SecureWipe() {
      i.Content = strings.Repeat("\x00", len(i.Content))
      // ... clear other sensitive fields
  }
  ```
- Add option to exclude by regex pattern (credit cards, SSN, etc.)

### 8.2 Encryption Enhancements

**Current State**: AES-256-GCM encryption at rest.

**Suggested Improvements**:
- Add optional password protection for history
- Implement key derivation from user password
- Add encryption key rotation support

---

## 9. Developer Experience

### 9.1 Build and Release Improvements

**Current State**: Basic build.sh script with Fyne packaging.

**Suggested Improvements**:
- Add Makefile for common tasks
- Add Docker build support for cross-compilation
- Add CI/CD configuration (GitHub Actions)
- Implement semantic versioning
- Add build version information in app

### 9.2 Documentation

**Current State**: README.md with setup and usage instructions.

**Suggested Improvements**:
- Add API documentation for internal packages
- Add CONTRIBUTING.md with development guidelines
- Add architecture documentation
- Add TROUBLESHOOTING.md with common issues

### 9.3 Profiling and Debugging

**Current State**: Benchmark tests exist.

**Suggested Improvements**:
- Add pprof endpoints for runtime profiling
- Add debug mode with verbose logging
- Add diagnostic commands in CLI:
  ```bash
  fyclip --diagnose  # Export diagnostics
  fyclip --profile   # Start with profiling enabled
  ```

---

## 10. Future Considerations

### 10.1 Cloud Sync (Long-term)

- Optional cloud backup/sync feature
- End-to-end encrypted sync
- Cross-device clipboard sharing

### 10.2 Plugin System (Long-term)

- Plugin API for extensions
- Community-contributed plugins
- Scripting support (Lua, JavaScript)

### 10.3 Mobile Companion (Long-term)

- iOS/Android companion app
- QR code sync between desktop and mobile
- Secure channel for sensitive data

---

## Summary

This project is well-structured with good performance optimizations already in place. The main areas for improvement are:

1. **Testing** - Add unit tests for core functionality
2. **Data Structures** - Add hash maps for O(1) operations
3. **Configuration** - Expand settings system
4. **Features** - Implement planned enhancements (global hotkey, snippets, exclusions)
5. **Code Quality** - Improve error handling and logging
6. **Platform Integration** - Enhance platform-specific features

The existing codebase shows good practices:
- Modular architecture
- Thread-safe implementations
- Debounced updates
- Encrypted storage
- Platform abstraction

These improvements would enhance functionality while maintaining the project's clean architecture and performance characteristics.
