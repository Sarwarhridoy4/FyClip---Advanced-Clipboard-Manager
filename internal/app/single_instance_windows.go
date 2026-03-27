//go:build windows

// File: internal/app/single_instance_windows.go
package app

import (
	"syscall"
	"unsafe"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess       = kernel32.NewProc("OpenProcess")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procGetExitCodeProcess = kernel32.NewProc("GetExitCodeProcess")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	STILL_ACTIVE              = 259
)

// isProcessRunning checks if a process with the given PID is running
// On Windows, we use OpenProcess and GetExitCodeProcess to check process existence
func isProcessRunning(pid int) bool {
	// Try to open the process with query information permission
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)
	
	if handle == 0 {
		// Process doesn't exist or we can't access it
		return false
	}
	
	defer procCloseHandle.Call(handle)
	
	// Get the exit code
	var exitCode uint32
	ret, _, _ := procGetExitCodeProcess.Call(
		handle,
		uintptr(unsafe.Pointer(&exitCode)),
	)
	
	if ret == 0 {
		// Failed to get exit code
		return false
	}
	
	// If exit code is STILL_ACTIVE, the process is still running
	return exitCode == STILL_ACTIVE
}
