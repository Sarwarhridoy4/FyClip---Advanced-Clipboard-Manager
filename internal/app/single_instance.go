// File: internal/app/single_instance.go
package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// singleInstanceLock implements a cross-platform single instance lock
type singleInstanceLock struct {
	lockFile *os.File
}

// lockFileName is the name of the lock file
const lockFileName = ".fyclip.lock"

// NewSingleInstanceLock creates a new single instance lock
// Returns the lock if successful, nil if another instance is already running
func NewSingleInstanceLock() (*singleInstanceLock, error) {
	lock := &singleInstanceLock{}

	// Get the appropriate lock file path based on OS
	lockPath, err := getLockFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get lock file path: %w", err)
	}

	log.Printf("Single instance: checking lock file at %s", lockPath)

	// Create the lock file with exclusive access
	// O_EXCL ensures atomic creation - fails if file already exists
	// O_RDWR allows us to write our PID and read to check
	lockFile, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			// Check if the existing process is still running
			if isPreviousInstanceRunning(lockPath) {
				return nil, fmt.Errorf("another instance is already running")
			}
			// Stale lock file - remove it and try again
			if err := os.Remove(lockPath); err != nil {
				return nil, fmt.Errorf("failed to remove stale lock file: %w", err)
			}
			// Try again
			lockFile, err = os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				return nil, fmt.Errorf("failed to create lock file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create lock file: %w", err)
		}
	}

	// Write our PID to the lock file
	pid := os.Getpid()
	_, err = lockFile.WriteString(fmt.Sprintf("%d\n", pid))
	if err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		return nil, fmt.Errorf("failed to write PID to lock file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := lockFile.Sync(); err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		return nil, fmt.Errorf("failed to sync lock file: %w", err)
	}

	lock.lockFile = lockFile
	return lock, nil
}

// getLockFilePath returns the path to the lock file
func getLockFilePath() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("APPDATA")
		}
	case "darwin":
		home := os.Getenv("HOME")
		if home != "" {
			baseDir = filepath.Join(home, "Library", "Application Support")
		}
	default: // linux
		// Use XDG_DATA_HOME if set, otherwise ~/.local/share
		baseDir = os.Getenv("XDG_DATA_HOME")
		if baseDir == "" {
			home := os.Getenv("HOME")
			if home != "" {
				baseDir = filepath.Join(home, ".local", "share")
			}
		}
	}

	if baseDir == "" {
		// Fallback to current directory (not ideal but better than nothing)
		baseDir = "."
	}

	// Create the directory if it doesn't exist
	appDir := filepath.Join(baseDir, "FyClip")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}

	return filepath.Join(appDir, lockFileName), nil
}

// isPreviousInstanceRunning checks if a previous instance is still running
func isPreviousInstanceRunning(lockPath string) bool {
	// Try to open the lock file
	file, err := os.Open(lockPath)
	if err != nil {
		// Can't read the file, assume no previous instance
		return false
	}
	defer file.Close()

	// Read the PID
	var pid int
	_, err = fmt.Fscan(file, &pid)
	if err != nil {
		// Can't read PID, assume no previous instance
		return false
	}

	if pid <= 0 {
		return false
	}

	// Check if the process is still running using platform-specific implementation
	return isProcessRunning(pid)
}

// Release releases the single instance lock
func (l *singleInstanceLock) Release() {
	if l.lockFile == nil {
		return
	}

	// Get the lock file path before closing
	lockPath := l.lockFile.Name()

	// Close and remove the lock file
	l.lockFile.Close()
	os.Remove(lockPath)
	l.lockFile = nil
}
