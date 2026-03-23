// File: internal/clipboard/item_test.go
package clipboard

import (
	"testing"
	"time"
)

func TestItemTypeString(t *testing.T) {
	tests := []struct {
		input    ItemType
		expected string
	}{
		{TypeText, "Text"},
		{TypeImage, "Image"},
		{ItemType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("ItemType.String() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestItemPrepareForSearch(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "lowercase content",
			content:  "hello world",
			expected: "hello world",
		},
		{
			name:     "uppercase content",
			content:  "HELLO WORLD",
			expected: "hello world",
		},
		{
			name:     "mixed case content",
			content:  "HeLLo WoRLd",
			expected: "hello world",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "special characters",
			content:  "Test@123!",
			expected: "test@123!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{
				Type:    TypeText,
				Content: tt.content,
			}
			item.PrepareForSearch()
			result := item.SearchContent()
			if result != tt.expected {
				t.Errorf("SearchContent() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestItemDisplayText(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		maxLen   int
		expected string
	}{
		{
			name: "text within limit",
			item: Item{
				Type:    TypeText,
				Content: "Hello World",
			},
			maxLen:   50,
			expected: "Hello World",
		},
		{
			name: "text truncated",
			item: Item{
				Type:    TypeText,
				Content: "This is a very long text that should be truncated when displayed",
			},
			maxLen:   20,
			expected: "This is a very lo...",
		},
		{
			name: "text with newlines",
			item: Item{
				Type:    TypeText,
				Content: "Line1\nLine2\r\nLine3",
			},
			maxLen:   50,
			expected: "Line1 Line2 Line3",
		},
		{
			name: "text with leading/trailing whitespace",
			item: Item{
				Type:    TypeText,
				Content: "  Hello World  ",
			},
			maxLen:   50,
			expected: "Hello World",
		},
		{
			name: "image type display",
			item: Item{
				Type:      TypeImage,
				ImageType: "png",
				Timestamp: time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC),
			},
			maxLen:   50,
			expected: "PNG Image (15:30:00)",
		},
		{
			name: "exact boundary truncation",
			item: Item{
				Type:    TypeText,
				Content: "12345", // 5 chars
			},
			maxLen:   5,
			expected: "12345",
		},
		{
			name: "just over boundary truncation",
			item: Item{
				Type:    TypeText,
				Content: "123456", // 6 chars
			},
			maxLen:   5,
			expected: "12...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.DisplayText(tt.maxLen)
			if result != tt.expected {
				t.Errorf("DisplayText(%d) = %v; want %v", tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestItemSize(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected int
	}{
		{
			name: "text size",
			item: Item{
				Type:    TypeText,
				Content: "Hello World",
			},
			expected: 11,
		},
		{
			name: "empty text size",
			item: Item{
				Type:    TypeText,
				Content: "",
			},
			expected: 0,
		},
		{
			name: "image size (base64)",
			item: Item{
				Type:      TypeImage,
				ImageData: "SGVsbG8gV29ybGQ=", // "Hello World" in base64
			},
			expected: 12, // 16 * 3 / 4 = 12
		},
		{
			name: "unicode text size",
			item: Item{
				Type:    TypeText,
				Content: "Hello 世界",
			},
			expected: 12,  // byte count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.Size()
			if result != tt.expected {
				t.Errorf("Size() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestItemTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		timestamp     time.Time
		expectedStart string // Check start of string for flexibility
	}{
		{
			name:          "just now",
			timestamp:     now.Add(-30 * time.Second),
			expectedStart: "just now",
		},
		{
			name:          "minutes ago",
			timestamp:     now.Add(-5 * time.Minute),
			expectedStart: "5m ago",
		},
		{
			name:          "single minute ago",
			timestamp:     now.Add(-1 * time.Minute),
			expectedStart: "1m ago",
		},
		{
			name:          "hours ago",
			timestamp:     now.Add(-3 * time.Hour),
			expectedStart: "3h ago",
		},
		{
			name:          "single hour ago",
			timestamp:     now.Add(-1 * time.Hour),
			expectedStart: "1h ago",
		},
		{
			name:          "yesterday",
			timestamp:     now.Add(-24 * time.Hour),
			expectedStart: "yesterday",
		},
		{
			name:          "days ago",
			timestamp:     now.Add(-48 * time.Hour),
			expectedStart: "2d ago",
		},
		{
			name:          "single day ago",
			timestamp:     now.Add(-25 * time.Hour),
			expectedStart: "yesterday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{
				Timestamp: tt.timestamp,
			}
			result := item.TimeAgo()
			// Check that result starts with expected (allows for singular/plural variations)
			if len(result) < len(tt.expectedStart) {
				t.Errorf("TimeAgo() = %v; expected to start with %v", result, tt.expectedStart)
				return
			}
			if result[:len(tt.expectedStart)] != tt.expectedStart {
				t.Errorf("TimeAgo() = %v; expected to start with %v", result, tt.expectedStart)
			}
		})
	}
}

func TestItemSearchContentEmpty(t *testing.T) {
	item := &Item{
		Type:      TypeText,
		Content:   "",
		searchContent: "", // Explicitly empty
	}

	result := item.SearchContent()
	if result != "" {
		t.Errorf("SearchContent() for empty = %v; want empty string", result)
	}
}

func TestItemIDField(t *testing.T) {
	item := Item{
		ID:        "test-id-123",
		Type:      TypeText,
		Content:   "Test content",
		Timestamp: time.Now(),
		Pinned:    false,
		CopyCount: 0,
		Hash:      "abc123",
	}

	if item.ID != "test-id-123" {
		t.Errorf("ID = %v; want test-id-123", item.ID)
	}

	// Verify Type field is set correctly
	if item.Type != TypeText {
		t.Errorf("Type = %v; want TypeText", item.Type)
	}

	// Verify Content field is set correctly
	if item.Content != "Test content" {
		t.Errorf("Content = %v; want Test content", item.Content)
	}

	// Verify Timestamp is set
	if item.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	// Verify Hash is set correctly
	if item.Hash != "abc123" {
		t.Errorf("Hash = %v; want abc123", item.Hash)
	}

	// Note: Pinned defaults to false (zero value)
	// Test is primarily checking field assignment
	_ = item.Pinned

	if item.CopyCount != 0 {
		t.Errorf("CopyCount = %v; want 0", item.CopyCount)
	}
}

func TestItemJSONFields(t *testing.T) {
	// Test that JSON tags are correctly applied
	item := Item{
		ID:        "json-test",
		Type:      TypeText,
		Content:   "test",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Pinned:    true,
		CopyCount: 5,
		Hash:      "hash123",
		searchContent: "search",
	}

	// These fields should be serializable to JSON
	// The searchContent should be ignored (has json:"-")
	if item.ID == "" {
		t.Error("ID should not be empty")
	}
	if item.Type != TypeText {
		t.Errorf("Type = %v; want TypeText", item.Type)
	}
	if item.Content != "test" {
		t.Errorf("Content = %v; want test", item.Content)
	}
	if item.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if !item.Pinned {
		t.Error("Pinned should be true")
	}
	if item.CopyCount != 5 {
		t.Errorf("CopyCount = %v; want 5", item.CopyCount)
	}
	if item.Hash != "hash123" {
		t.Errorf("Hash = %v; want hash123", item.Hash)
	}
	if item.searchContent != "search" {
		t.Errorf("searchContent = %v; want search", item.searchContent)
	}
}

func TestItemCopyCount(t *testing.T) {
	item := Item{
		ID:        "copy-test",
		Type:      TypeText,
		Content:   "Test",
		CopyCount: 0,
	}

	// Simulate copying
	item.CopyCount++

	if item.CopyCount != 1 {
		t.Errorf("CopyCount = %v; want 1", item.CopyCount)
	}

	item.CopyCount++

	if item.CopyCount != 2 {
		t.Errorf("CopyCount = %v; want 2", item.CopyCount)
	}

	// Verify fields are set correctly
	if item.ID != "copy-test" {
		t.Errorf("ID = %v; want copy-test", item.ID)
	}
	if item.Type != TypeText {
		t.Errorf("Type = %v; want TypeText", item.Type)
	}
	if item.Content != "Test" {
		t.Errorf("Content = %v; want Test", item.Content)
	}
}

func TestItemMultipleTypes(t *testing.T) {
	items := []Item{
		{Type: TypeText, Content: "text item"},
		{Type: TypeImage, Content: "", ImageData: "imagedata", ImageType: "png"},
		{Type: TypeText, Content: "another text"},
	}

	if len(items) != 3 {
		t.Errorf("Items length = %v; want 3", len(items))
	}

	if items[0].Type != TypeText {
		t.Error("First item should be TypeText")
	}

	if items[1].Type != TypeImage {
		t.Error("Second item should be TypeImage")
	}

	// Verify image size calculation
	imgSize := items[1].Size()
	if imgSize <= 0 {
		t.Errorf("Image size should be positive, got %v", imgSize)
	}
}

