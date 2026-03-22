# FyClip Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Changed
- Improved clipboard monitoring performance
- Enhanced search functionality
- Better item deduplication

### Fixed
- Various bug fixes and improvements

---

## [Previous Versions]

See the [release tags](https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/tags) for version history.
