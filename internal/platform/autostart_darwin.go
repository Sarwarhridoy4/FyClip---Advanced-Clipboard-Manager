// File: internal/platform/autostart_darwin.go
//go:build darwin

package platform

import (
	"fmt"
	"os"
)

// enableDarwin creates a plist file for autostart on macOS
func (as *AutoStart) enableDarwin() error {
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