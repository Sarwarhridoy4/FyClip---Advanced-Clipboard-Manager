// File: internal/platform/autostart.go
package platform

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// AutoStart manages application autostart
type AutoStart struct {
	execPath string
	filePath string
}

// NewAutoStart creates a new autostart manager
func NewAutoStart(execPath string) *AutoStart {
	as := &AutoStart{
		execPath: execPath,
	}
	as.filePath = as.getAutoStartPath()
	return as
}

// getAutoStartPath returns the platform-specific autostart file path
func (as *AutoStart) getAutoStartPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Warning: Could not get current user: %v", err)
		return ""
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(usr.HomeDir, ".config", "autostart", "fyclip.desktop")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			log.Printf("Warning: APPDATA environment variable not set")
			return ""
		}
		return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "fyclip.lnk")
	case "darwin":
		return filepath.Join(usr.HomeDir, "Library", "LaunchAgents", "com.fyclip.plist")
	}

	return ""
}

// IsEnabled checks if autostart is enabled
func (as *AutoStart) IsEnabled() bool {
	if as.filePath == "" {
		return false
	}

	_, err := os.Stat(as.filePath)
	return err == nil
}

// Enable enables autostart
func (as *AutoStart) Enable() error {
	if as.filePath == "" {
		return fmt.Errorf("autostart not supported on this platform")
	}

	if as.execPath == "" {
		return fmt.Errorf("executable path not set")
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(as.filePath), 0755); err != nil {
		return fmt.Errorf("failed to create autostart directory: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return as.enableLinux()
	case "windows":
		return as.enableWindows()
	case "darwin":
		return as.enableDarwin()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (as *AutoStart) enableDarwin() error {
	panic("unimplemented")
}

func (as *AutoStart) enableLinux() error {
	panic("unimplemented")
}

// Disable disables autostart
func (as *AutoStart) Disable() error {
	if as.filePath == "" {
		return fmt.Errorf("autostart not supported on this platform")
	}

	if err := os.Remove(as.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove autostart file: %w", err)
	}

	return nil
}
