// File: internal/clipboard/item.go
package clipboard

import (
	"fmt"
	"strings"
	"time"
)

// ItemType represents the type of clipboard content
type ItemType int

const (
	TypeText ItemType = iota
	TypeImage
)

func (t ItemType) String() string {
	switch t {
	case TypeText:
		return "Text"
	case TypeImage:
		return "Image"
	default:
		return "Unknown"
	}
}

// Item represents a clipboard history item
type Item struct {
	ID        string    `json:"id"`
	Type      ItemType  `json:"type"`
	Content   string    `json:"content"`
	ImageData string    `json:"image_data,omitempty"`
	ImageType string    `json:"image_type,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Pinned    bool      `json:"pinned"`
	Hash      string    `json:"hash,omitempty"`
}

// DisplayText returns a truncated version of the content for display
func (i *Item) DisplayText(maxLen int) string {
	if i.Type == TypeImage {
		return fmt.Sprintf("%s Image (%s)", 
			strings.ToUpper(i.ImageType), 
			i.Timestamp.Format("15:04:05"))
	}
	
	text := strings.ReplaceAll(i.Content, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimSpace(text)
	
	if len(text) > maxLen {
		text = text[:maxLen-3] + "..."
	}
	return text
}

// Size returns approximate size in bytes
func (i *Item) Size() int {
	if i.Type == TypeImage {
		return len(i.ImageData) * 3 / 4 // base64 overhead
	}
	return len(i.Content)
}

// TimeAgo returns a human-readable time difference
func (i *Item) TimeAgo() string {
	diff := time.Since(i.Timestamp)
	
	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "yesterday"
	}
	return fmt.Sprintf("%dd ago", days)
}