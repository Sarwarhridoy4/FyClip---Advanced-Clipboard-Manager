//go:build !windows

// File: internal/app/single_instance_unix.go
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// isProcessRunning checks if a process with the given PID is running
// On Unix systems, we use syscall.Kill with signal 0 to check process existence
func isProcessRunning(pid int, expectedExecPath string) bool {
	err := syscall.Kill(pid, 0)
	if err != nil {
		return err != syscall.ESRCH
	}

	runningExecPath, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return false
	}

	if runningExecPath == "" {
		return false
	}

	if expectedExecPath != "" {
		if sameExecutable(runningExecPath, expectedExecPath) {
			return true
		}
		return false
	}

	currentExecPath, err := os.Executable()
	if err == nil && currentExecPath != "" && sameExecutable(runningExecPath, currentExecPath) {
		return true
	}

	return looksLikeFyClipProcess(runningExecPath)
}

func sameExecutable(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func looksLikeFyClipProcess(execPath string) bool {
	base := filepath.Base(execPath)
	return base == "fyclip" || base == "FyClip" || base == "FyClip---Advanced-Clipboard-Manager"
}
