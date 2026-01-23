// File: internal/platform/autostart.go
package platform

import "fmt"

// AutoStart manages application autostart
type AutoStart struct {
	execPath string
	filePath string
}

// NewAutoStart creates a new autostart manager
func NewAutoStart(execPath string) *AutoStart {
	return &AutoStart{
		execPath: execPath,
		filePath: getAutoStartPath(),
	}
}

// IsEnabled checks if autostart is enabled
func (as *AutoStart) IsEnabled() bool {
	if as.filePath == "" {
		return false
	}
	return fileExists(as.filePath)
}

// Enable enables autostart
func (as *AutoStart) Enable() error {
	if as.execPath == "" {
		return fmt.Errorf("executable path not set")
	}
	if as.filePath == "" {
		return fmt.Errorf("autostart not supported on this platform")
	}
	return as.enable()
}

// Disable disables autostart
func (as *AutoStart) Disable() error {
	if as.filePath == "" {
		return fmt.Errorf("autostart not supported on this platform")
	}
	return as.disable()
}
