// File: internal/platform/autostart_linux.go
//go:build linux

package platform

import (
	"fmt"
	"os"
)

// enableLinux creates a .desktop file for autostart on Linux
func (as *AutoStart) enableLinux() error {
	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Name=FyClip
Comment=Clipboard Manager
`, as.execPath)

	return os.WriteFile(as.filePath, []byte(content), 0644)
}