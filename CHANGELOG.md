# FyClip Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

---

## [Previous Versions]

See the [release tags](https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/tags) for version history.
