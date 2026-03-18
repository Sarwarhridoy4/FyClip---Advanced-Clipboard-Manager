// File: internal/platform/autostart_test.go
package platform

import (
	"testing"
)

// TestAutoStartNew tests creating a new AutoStart instance
func TestAutoStartNew(t *testing.T) {
	// Test with empty path (should not panic)
	as := NewAutoStart("")
	if as == nil {
		t.Fatal("NewAutoStart returned nil")
	}
	
	// Test with valid path
	as = NewAutoStart("/usr/bin/test")
	if as == nil {
		t.Fatal("NewAutoStart returned nil")
	}
}

// TestAutoStartIsEnabled tests checking autostart status
func TestAutoStartIsEnabled(t *testing.T) {
	as := NewAutoStart("/usr/bin/test")
	
	// Should not panic - just check if it returns a value
	// The actual result depends on system state
	_ = as.IsEnabled()
}

// TestAutoStartEnable tests enabling autostart
func TestAutoStartEnable(t *testing.T) {
	as := NewAutoStart("/usr/bin/test")
	
	// Enable may fail (e.g., no permissions)
	// But should not panic
	err := as.Enable()
	if err != nil {
		t.Logf("Enable error (may be expected): %v", err)
	}
}

// TestAutoStartDisable tests disabling autostart
func TestAutoStartDisable(t *testing.T) {
	as := NewAutoStart("/usr/bin/test")
	
	// Disable may fail if not enabled
	// But should not panic
	err := as.Disable()
	if err != nil {
		t.Logf("Disable error (may be expected): %v", err)
	}
}
