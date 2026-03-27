# FyClip Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.2.0] - 2026-03-25

### Added
- **Single Instance Protection**: Prevents multiple FyClip instances from running simultaneously
  - Uses lock file mechanism to detect existing instances
  - Lock file location: `~/.local/share/FyClip/.fyclip.lock` (Linux), `%LOCALAPPDATA%/FyClip/` (Windows), `~/Library/Application Support/FyClip/` (macOS)
  - Automatically handles stale lock files from crashed instances
  - Shows system notification when another instance is already running
    - Linux: Critical urgency notification via `shownotification`
    - Windows: PowerShell toast notification
    - macOS: `osascript` notification

### Performance
- **Image Thumbnails**: Added 150x150px thumbnails for efficient memory usage
  - List view displays thumbnails instead of full images
  - Preview pane shows full images when requested
  - Generates thumbnails automatically for captured images
  - Backward compatibility: auto-generates thumbnails for existing items on load
  - Expected: 50-80% memory reduction for image-heavy clipboard history
- **Storage Compression**: Added gzip compression before encryption
  - Compresses JSON data before AES-256-GCM encryption
  - Decompresses after decryption on load
  - Backward compatibility: handles both compressed and uncompressed data
  - Expected: 30-60% reduction in storage file size
- **JSON Optimization**: Removed indentation whitespace from saved JSON
  - Changed from `json.MarshalIndent()` to `json.Marshal()`
  - Expected: ~10-15% additional reduction in storage file size
- **Channel Buffer Optimization**: Reduced update channel buffer from 100 to 10
  - Updates are debounced anyway, so smaller buffer is sufficient
  - Minor memory reduction for channel buffers

## [2.1.3] - 2026-03-24

### Added
- **Auto Update Feature**: Check for and install application updates automatically
  - GitHub release detection for checking latest version
  - Automatic download of platform-specific packages
  - Terminal commands: `fyclip --check-update` and `fyclip --update`
  - UI integration: Help menu "Check for Updates" option
  - Feature guide documentation
  - Supports Linux (.deb, .AppImage), Windows (.exe, .msi), macOS (.dmg)

## [2.1.2] - 2026-03-24

### Added
- **Theme Support**: Light, Dark, and System theme modes with centered popup selection
  - Toggle between Light, Dark, and System (follows OS) themes
  - Theme selection popup centered on window for better UX

### Fixed
- **HTML Preview**: Auto-detect HTML content and display as code block
  - Added fast HTML detection in clipboard monitor
  - Any content starting with `<` followed by a letter is detected as HTML
  - HTML content displays as code block in preview pane using `showCode()` function
- **Unused Function Warning**: Exported `clearRegexCache` to `ClearRegexCache` for test usage
- **Race Condition**: Fixed thread-safety issue in `searchWithRegex` by using single lock with defer
- **Unused Theme Select**: Removed unused `onThemeSelect` method and incomplete dropdown code
- **Theme Popup Position**: Centered theme selection popup menu on the window

### Performance
- **Object Pool Integration**: Added sync.Pool for Item reuse to reduce GC pressure
- **Regex Cache**: Compiled regex patterns cached for faster repeated searches
- **Fuzzy Search Optimization**: Optimized subsequence matching with reduced allocations

## [2.1.1] - 2026-03-24

### Added
- **Bulk Operations**: Multi-select clipboard items with checkboxes
  - Toggle selection mode via toolbar button
  - Bulk delete (skips pinned items)
  - Bulk pin/unpin functionality
  - Select all/none options

- **Keyboard Navigation**: Enhanced keyboard shortcuts
  - Arrow keys for navigation
  - Enter to copy, Delete to delete
  - Space to toggle selection or copy
  - Escape to exit selection mode
  - Home/End to jump to top/bottom
  - F1 for quick panel

- **Smart Categories & Tags**: Automatic content categorization
  - Auto-detection based on content patterns:
    - URLs → "Links"
    - Emails → "Contacts"
    - Phone numbers → "Contacts"
    - Code snippets → "Code"
    - File paths → "Files"
    - JSON → "Data"
    - Images → "Images"
    - HTML → "Web"
  - Manual tag support (AddTag, RemoveTag, HasTag)

- **Quick Panel**: Global hotkey quick access popup (Ctrl+Shift+V)

- **Snippets/Templates**: Text templates with variables
  - Support for {{date}}, {{time}}, {{datetime}}, {{year}}, {{month}}, {{day}}, {{clipboard}}

- **Pattern Exclusion**: Regex, app, and size-based content filtering

- **Hash Maps**: O(1) duplicate detection and item lookup

- **Encrypted Backup**: Password-protected backup with AES-256-GCM

- **Rich Text/HTML**: Capture and preserve HTML clipboard content

- **File History**: Track files copied from file manager

- **Enhanced Search**: Regex, case-sensitive, and fuzzy matching

- **Sensitive Data Detection**: Auto-detect credit cards, SSN, API keys

- **Structured Logging**: slog-based logging with file rotation

- **Graceful Shutdown**: Context-based shutdown with hooks

- **System Tray**: Recent items submenu, Clear History action

- **Preview Enhancements**: JSON pretty-printing, file info display

### Changed
- Improved clipboard monitoring performance
- Enhanced search functionality
- Better item deduplication

### Fixed
- Fixed unused write warnings in test files (improved test coverage)
- Added missing field assertions in TestItemIDField, TestItemJSONFields, and TestItemCopyCount
- Resolved import metadata issue with internal/ui package (gopls cache)
- Updated .gitignore with proper build artifact patterns
- Synced dependencies in go.sum

### Performance
- **Object Pool Integration**: Added sync.Pool for Item reuse to reduce GC pressure
- **Regex Cache**: Compiled regex patterns cached for faster repeated searches
- **Fuzzy Search Optimization**: Optimized subsequence matching with reduced allocations

---

## [Previous Versions]

See the [release tags](https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/tags) for version history.
