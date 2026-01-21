// File: internal/platform/autostart_windows.go
//go:build windows

package platform

import (
	"fmt"
	"os/exec"
)

// enableWindows creates a shortcut in the Startup folder for Windows
func (as *AutoStart) enableWindows() error {
	cmd := fmt.Sprintf(`$ws = New-Object -ComObject WScript.Shell;
$lnk = $ws.CreateShortcut("%s");
$lnk.TargetPath = "%s";
$lnk.Save()`, as.filePath, as.execPath)

	return exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", cmd).Run()
}