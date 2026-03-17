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
	// Match the application's desktop id (Fyne app ID) so GNOME can associate the window with this entry.
	return filepath.Join(home, ".config", "autostart", "com.sarwar.fyclip.desktop")
}

func (as *AutoStart) enable() error {
	if err := os.MkdirAll(filepath.Dir(as.filePath), 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Exec=%s
Icon=com.sarwar.fyclip
StartupWMClass=com.sarwar.fyclip
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
