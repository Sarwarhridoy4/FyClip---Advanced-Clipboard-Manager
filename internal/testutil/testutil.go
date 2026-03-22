// File: internal/testutil/testutil.go
package testutil

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// NewTestManager creates a manager for testing with a temporary storage path.
// It automatically handles cleanup after the test.
func NewTestManager(t *testing.T) *clipboard.Manager {
	t.Helper()

	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test.enc")

	manager, err := clipboard.NewManager(clipboard.Config{
		StoragePath: storagePath,
		OnError: func(err error) {
			t.Logf("Manager error: %v", err)
		},
		OnInfo: func(msg string) {
			t.Logf("Manager info: %s", msg)
		},
	})
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Ensure cleanup
	t.Cleanup(func() {
		if manager != nil {
			manager.Shutdown()
		}
	})

	return manager
}

// CreateTestItem creates a test item with default values.
func CreateTestItem(content string) clipboard.Item {
	return clipboard.Item{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      clipboard.TypeText,
		Content:   content,
		Timestamp: time.Now(),
		Pinned:    false,
		CopyCount: 0,
	}
}

// CreateTestImageItem creates a test image item.
func CreateTestImageItem(imageData, imageType string) clipboard.Item {
	return clipboard.Item{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:       clipboard.TypeImage,
		ImageData:  imageData,
		ImageType:  imageType,
		Timestamp:  time.Now(),
		Pinned:     false,
		CopyCount:  0,
	}
}

// CreateTestFileItem creates a test file item.
func CreateTestFileItem(name, path string, size int64) clipboard.Item {
	return clipboard.Item{
		ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:     clipboard.TypeFile,
		Content:  path,
		FileInfo: &clipboard.FileInfo{
			Name:    name,
			Path:    path,
			Size:    size,
			ModTime: time.Now(),
		},
		Timestamp: time.Now(),
		Pinned:    false,
		CopyCount: 0,
	}
}

// CreateTestHTMLItem creates a test HTML item.
func CreateTestHTMLItem(content, htmlContent string) clipboard.Item {
	return clipboard.Item{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:        clipboard.TypeHTML,
		Content:     content,
		HTMLContent: htmlContent,
		Timestamp:   time.Now(),
		Pinned:      false,
		CopyCount:   0,
	}
}

// CreateTestItems creates multiple test items.
func CreateTestItems(count int) []clipboard.Item {
	items := make([]clipboard.Item, count)
	for i := 0; i < count; i++ {
		items[i] = CreateTestItem(fmt.Sprintf("Test content %d", i))
	}
	return items
}

// AssertItemCount asserts that the manager has the expected number of items.
func AssertItemCount(t *testing.T, manager *clipboard.Manager, expected int) {
	t.Helper()

	items := manager.GetFiltered()
	if len(items) != expected {
		t.Errorf("Expected %d items, got %d", expected, len(items))
	}
}

// AssertItemContent asserts that an item has the expected content.
func AssertItemContent(t *testing.T, item *clipboard.Item, expectedContent string) {
	t.Helper()

	if item == nil {
		t.Fatal("Item is nil")
	}
	if item.Content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, item.Content)
	}
}

// AssertItemType asserts that an item has the expected type.
func AssertItemType(t *testing.T, item *clipboard.Item, expectedType clipboard.ItemType) {
	t.Helper()

	if item == nil {
		t.Fatal("Item is nil")
	}
	if item.Type != expectedType {
		t.Errorf("Expected type %v, got %v", expectedType, item.Type)
	}
}
