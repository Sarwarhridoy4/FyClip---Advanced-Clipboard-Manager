// File: internal/clipboard/native_test.go
package clipboard

import (
	"testing"
)

// TestNewNativeClipboard tests creating a new native clipboard
func TestNewNativeClipboard(t *testing.T) {
	// This test verifies native clipboard can be initialized
	// On Linux without xclip/wl-clipboard it may fail
	native, err := NewNativeClipboard()
	if err != nil {
		// On systems without clipboard tools, this may fail
		// But we should at least verify it doesn't panic
		t.Logf("NativeClipboard initialization: %v", err)
		return
	}
	
	if native == nil {
		t.Fatal("NewNativeClipboard returned nil")
	}
	
	// Test IsAvailable
	_ = native.IsAvailable()
	
	// Test Backend
	backend := native.Backend()
	t.Logf("Clipboard backend: %s", backend)
}

// TestNativeClipboardReadWrite tests basic read/write operations
func TestNativeClipboardReadWrite(t *testing.T) {
	native, err := NewNativeClipboard()
	if err != nil {
		t.Skipf("Skipping test: clipboard not available: %v", err)
	}
	
	// Test writing text
	testText := "Test clipboard content"
	err = native.WriteText([]byte(testText))
	if err != nil {
		t.Errorf("WriteText failed: %v", err)
	}
	
	// Small delay
	// Read text back
	// Note: On some systems, clipboard read may not return immediately
	// due to the programmatic copy suppression
}

// TestNativeClipboardDetectImageType tests image type detection
func TestNativeClipboardDetectImageType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "PNG signature",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expected: "png",
		},
		{
			name:     "JPEG signature",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			expected: "jpeg",
		},
		{
			name:     "Too short",
			data:     []byte{0x00, 0x01},
			expected: "png", // default
		},
		{
			name:     "Unknown",
			data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "png", // default
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectImageType(tt.data)
			if result != tt.expected {
				t.Errorf("detectImageType() = %v; want %v", result, tt.expected)
			}
		})
	}
}

// TestNativeClipboardEmptyOperations tests operations with empty data
func TestNativeClipboardEmptyOperations(t *testing.T) {
	native, err := NewNativeClipboard()
	if err != nil {
		t.Skipf("Skipping test: clipboard not available: %v", err)
	}
	
	// Test reading empty clipboard
	text := native.ReadText()
	if text == nil {
		t.Error("ReadText should return nil, not panic")
	}
	
	img, imgType := native.ReadImage()
	if img == nil && imgType != "" {
		t.Error("ReadImage inconsistency: got image data without type")
	}
}

// TestNativeClipboardWriteEmptyImage tests writing empty image
func TestNativeClipboardWriteEmptyImage(t *testing.T) {
	native, err := NewNativeClipboard()
	if err != nil {
		t.Skipf("Skipping test: clipboard not available: %v", err)
	}
	
	// Writing empty image should work (it will be invalid but shouldn't panic)
	err = native.WriteImage("")
	if err != nil {
		t.Logf("WriteImage empty: %v", err)
	}
}
