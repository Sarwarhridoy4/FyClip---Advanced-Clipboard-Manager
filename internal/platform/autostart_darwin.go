//go:build darwin

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
	return filepath.Join(
		home,
		"Library",
		"LaunchAgents",
		"com.fyclip.plist",
	)
}

func (as *AutoStart) enable() error {
	if err := os.MkdirAll(filepath.Dir(as.filePath), 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.fyclip</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>`, as.execPath)

	return os.WriteFile(as.filePath, []byte(content), 0644)
}

func (as *AutoStart) disable() error {
	return os.Remove(as.filePath)
}
