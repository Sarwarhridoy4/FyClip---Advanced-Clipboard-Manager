# Changelog

## [Unreleased]

### Added
- **Rich Text/HTML Support**: Added TypeHTML item type, HTMLContent field, and HTML clipboard reading/writing support
- **File Path History**: Added TypeFile item type with FileInfo struct for capturing copied files from file manager
- **Quick Panel**: Global hotkey quick panel for fast paste from recent history
- **Snippets/Templates**: Snippet management with template variables ({{date}}, {{time}}, {{clipboard}}, etc.)
- **Pattern Exclusion Rules**: Configurable regex, app, and size-based exclusion rules
- **Enhanced Search**: Regex search, case-sensitive search, and fuzzy matching support
- **Backup/Restore**: Full backup/restore feature with AES-256-GCM encryption, optional password protection, merge/replace options, and UI toolbar buttons
- **Toolbar 2-Column Layout**: Reorganized toolbar into two rows to reduce window width
- **Menu with Backup/Restore**: Added Edit menu with Copy, Delete, Search, Backup, and Restore menu items
- **Hash Maps for O(1) Operations**: Added hashIndexMap and idIndexMap for efficient duplicate detection
- **Comprehensive Configuration**: Config system with max_history_items, monitoring_interval, theme, etc.
- **Error Handling**: Custom ClipboardError type with ErrorCode categories
- **Structured Logging**: slog-based logging with levels and file rotation
- **Graceful Shutdown**: Context-based shutdown with timeout and shutdown hooks
- **Sensitive Data Handling**: Pattern detection for credit cards, SSN, API keys, private keys
- **Makefile**: Build automation for all platforms
- **Version CLI**: --version and --help command-line options
- **Preview Enhancements**: JSON pretty-printing, HTML preview, file info display
- **System Tray Improvements**: Recent items submenu, Clear History action
- **Code Block Preview**: Preview pane now detects code content and displays it in syntax-highlighted code blocks with language detection (Go, Python, JavaScript, TypeScript, Java, C/C++, Rust, HTML, CSS, SQL, Bash, YAML, JSON, Markdown)
- **Snippet Usage Instructions**: Added detailed snippet usage documentation in README and Features Guide dialog
- **Enhanced Snippets UI**: Added Add, Delete, and Use functionality for snippets with live preview of template expansion
- **System Snippets**: Added IsSystem field to mark default snippets that cannot be deleted. User-defined snippets can be deleted.
- **Snippet Expansion with Clipboard**: New ExpandSnippetWithClipboard method for better template variable handling

### Changed
- Updated build process with Makefile targets
- Improved preview pane for HTML and file types
- Fixed Pin/Unpin race condition: Added FindIndexByID method to avoid stale index issues when toggling pin

## 1.5.1 - 2026-02-11

### Fixed
- Markdown preview rendering (rich markdown no longer appears as plain text)
- History list type icon mapping for text/image entries
- Selection/preview lag by removing redundant refresh paths and caching preview renders
- Linux launcher metadata alignment for better Ubuntu dock icon matching

### Changed
- Reworked `build.sh` to use Fyne's official Linux packaging output as the source for Debian and AppImage builds
- Updated `FyneApp.toml` metadata (`Name`, `ID`, `Icon`, and Linux/BSD fields)
- Updated `readme.md` and `SETUP_GUIDE.md` with the current packaging workflow and requirements

## 1.5.0 - 2026-01-31

### Added
- Explicit Markdown support to preview pane
- Enhanced image save dialog with explicit format selection
- History encryption and clarified image saving options
- Debian and AppImage builds
- Windows packaging

### Fixed
- Autostart related issue
- Issues with action buttons malfunctioning after code modularization
- Image preview not working
- Security updates

### Changed
- Refactored image format determination in `onSaveImage`
- Code modularization
- Updated build script
- Updated readme and build process for Windows

### Dependencies
- Bumped `golang.org/x/net` in `go_modules` group
