//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func escapePowerShellString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	return s
}

func (as *AutoStart) enable() error {
	script := fmt.Sprintf(`$ws = New-Object -ComObject WScript.Shell
$lnk = $ws.CreateShortcut('%s')
$lnk.TargetPath = '%s'
$lnk.Save()
`, escapePowerShellString(as.filePath), escapePowerShellString(as.execPath))

	tmpFile, err := os.CreateTemp("", "fyclip-autostart-*.ps1")
	if err != nil {
		return fmt.Errorf("failed to create temp autostart script: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		return fmt.Errorf("failed to write autostart script: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close autostart script: %w", err)
	}

	return exec.Command(
		"powershell",
		"-NoProfile",
		"-WindowStyle",
		"Hidden",
		"-ExecutionPolicy",
		"Bypass",
		"-File",
		tmpFile.Name(),
	).Run()
}

func (as *AutoStart) disable() error {
	return os.Remove(as.filePath)
}
