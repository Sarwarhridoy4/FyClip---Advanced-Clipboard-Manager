# Changelog

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
