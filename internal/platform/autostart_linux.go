//go:build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

func getAutoStartPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "autostart", "fyclip.desktop")
}

func (as *AutoStart) enable() error {
	if err := os.MkdirAll(filepath.Dir(as.filePath), 0755); err != nil {
		return err
	}

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

func (as *AutoStart) disable() error {
	return os.Remove(as.filePath)
}
