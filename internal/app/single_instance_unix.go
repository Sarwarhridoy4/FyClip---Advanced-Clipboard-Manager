//go:build !windows

// File: internal/app/single_instance_unix.go
package app

import (
	"syscall"
)

// isProcessRunning checks if a process with the given PID is running
// On Unix systems, we use syscall.Kill with signal 0 to check process existence
func isProcessRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err != nil {
		// ESRCH means no such process - stale lock file
		// EPERM means process exists but we don't have permission to signal it
		return err != syscall.ESRCH
	}
	// Process is running
	return true
}
