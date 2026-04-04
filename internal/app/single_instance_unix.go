//go:build !windows

// File: internal/app/single_instance_unix.go
package app

import (
	"fmt"
	"os"
	"syscall"
)

// isProcessRunning checks if a process with the given PID is running
// On Unix systems, we use syscall.Kill with signal 0 to check process existence
func isProcessRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err != nil {
		return err != syscall.ESRCH
	}

	exePath, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return true
	}

	return len(exePath) > 0
}
