//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func getAutoStartPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return ""
	}
	return filepath.Join(
		appData,
		"Microsoft",
		"Windows",
		"Start Menu",
		"Programs",
		"Startup",
		"fyclip.lnk",
	)
}

func (as *AutoStart) enable() error {
	cmd := fmt.Sprintf(`
$ws = New-Object -ComObject WScript.Shell
$lnk = $ws.CreateShortcut("%s")
$lnk.TargetPath = "%s"
$lnk.Save()
`, as.filePath, as.execPath)

	return exec.Command(
		"powershell",
		"-NoProfile",
		"-WindowStyle",
		"Hidden",
		"-Command",
		cmd,
	).Run()
}

func (as *AutoStart) disable() error {
	return os.Remove(as.filePath)
}
