// File: internal/clipboard/item.go
package clipboard

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ItemType represents the type of clipboard content
type ItemType int

const (
	TypeText ItemType = iota
	TypeImage
	TypeHTML
	TypeFile
)

func (t ItemType) String() string {
	switch t {
	case TypeText:
		return "Text"
	case TypeImage:
		return "Image"
	case TypeHTML:
		return "HTML"
	case TypeFile:
		return "File"
	default:
		return "Unknown"
	}
}

// FileInfo represents metadata for file clipboard items
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	IsDirectory bool      `json:"is_directory"`
}

// Item represents a clipboard history item
type Item struct {
	ID           string    `json:"id"`
	Type         ItemType  `json:"type"`
	Content      string    `json:"content"`
	ImageData    string    `json:"image_data,omitempty"`
	ImageType    string    `json:"image_type,omitempty"`
	HTMLContent  string    `json:"html_content,omitempty"`
	FileInfo     *FileInfo `json:"file_info,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Pinned       bool      `json:"pinned"`
	CopyCount    int       `json:"copy_count,omitempty"`
	Hash         string    `json:"hash,omitempty"`
	Category     string    `json:"category,omitempty"`
	Tags         []string  `json:"tags,omitempty"`

	searchContent string `json:"-"`
}

// PrepareForSearch builds cached normalized text used for filtering.
func (i *Item) PrepareForSearch() {
	i.searchContent = strings.ToLower(i.Content)
}

// SearchContent returns cached normalized text used for filtering.
func (i *Item) SearchContent() string {
	return i.searchContent
}

// DisplayText returns a truncated version of the content for display
func (i *Item) DisplayText(maxLen int) string {
	if i.Type == TypeImage {
		return fmt.Sprintf("%s Image (%s)",
			strings.ToUpper(i.ImageType),
			i.Timestamp.Format("15:04:05"))
	}

	if i.Type == TypeFile {
		if i.FileInfo != nil {
			return fmt.Sprintf("📁 %s (%s)", i.FileInfo.Name, formatFileSize(i.FileInfo.Size))
		}
		return fmt.Sprintf("File: %s", i.Timestamp.Format("15:04:05"))
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
	if i.Type == TypeFile && i.FileInfo != nil {
		return int(i.FileInfo.Size)
	}
	return len(i.Content)
}

// HasHTML returns true if the item has HTML content
func (i *Item) HasHTML() bool {
	return i.Type == TypeHTML && i.HTMLContent != ""
}

// IsFile returns true if the item is a file
func (i *Item) IsFile() bool {
	return i.Type == TypeFile && i.FileInfo != nil
}

// GetDisplayContent returns the appropriate content for display based on type
func (i *Item) GetDisplayContent() string {
	if i.Type == TypeHTML && i.HTMLContent != "" {
		// Return plain text version of HTML
		return stripHTMLTags(i.HTMLContent)
	}
	return i.Content
}

// stripHTMLTags removes HTML tags from content
func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), suffixes[exp])
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

// AutoDetectCategory detects and sets the category based on content patterns
func (i *Item) AutoDetectCategory() {
	if i.Category != "" {
		return // Already has a category
	}

	content := i.Content

	// URL detection
	if isURL(content) {
		i.Category = "Links"
		return
	}

	// Email detection
	if isEmail(content) {
		i.Category = "Contacts"
		return
	}

	// Phone number detection
	if isPhoneNumber(content) {
		i.Category = "Contacts"
		return
	}

	// Code detection
	if isCodeSnippet(content) {
		i.Category = "Code"
		return
	}

	// File path detection
	if isContentFilePath(content) {
		i.Category = "Files"
		return
	}

	// JSON detection
	if isJSON(content) {
		i.Category = "Data"
		return
	}

	// Default category based on type
	switch i.Type {
	case TypeImage:
		i.Category = "Images"
	case TypeFile:
		i.Category = "Files"
	case TypeHTML:
		i.Category = "Web"
	default:
		i.Category = "Text"
	}
}

// AddTag adds a tag to the item
func (i *Item) AddTag(tag string) {
	for _, t := range i.Tags {
		if t == tag {
			return // Tag already exists
		}
	}
	i.Tags = append(i.Tags, tag)
}

// RemoveTag removes a tag from the item
func (i *Item) RemoveTag(tag string) {
	newTags := make([]string, 0)
	for _, t := range i.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	i.Tags = newTags
}

// HasTag returns true if the item has the specified tag
func (i *Item) HasTag(tag string) bool {
	for _, t := range i.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// isURL checks if the content is a URL
func isURL(content string) bool {
	content = strings.TrimSpace(content)
	return strings.HasPrefix(content, "http://") ||
		strings.HasPrefix(content, "https://") ||
		strings.HasPrefix(content, "ftp://") ||
		strings.HasPrefix(content, "file://")
}

// isEmail checks if the content is an email address
func isEmail(content string) bool {
	content = strings.TrimSpace(content)
	match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, content)
	return match
}

// isPhoneNumber checks if the content is a phone number
func isPhoneNumber(content string) bool {
	content = strings.TrimSpace(content)
	// Match common phone number formats
	match, _ := regexp.MatchString(`^[+]?[0-9\s()-]{7,20}$`, content)
	return match
}

// isCodeSnippet checks if the content looks like code
func isCodeSnippet(content string) bool {
	// Common code patterns
	codeIndicators := []string{"func ", "function ", "def ", "class ", "const ", "let ", "var ",
		"import ", "export ", "package ", "#include", "public ", "private ",
		"if (", "for (", "while (", "switch (", "=> {", "-> {",
		"{}", "};", "();", "[]", "<>", "<!--", "//", "/*", "#"}

	count := 0
	for _, indicator := range codeIndicators {
		if strings.Contains(content, indicator) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

// isContentFilePath checks if the content looks like a file path
func isContentFilePath(content string) bool {
	content = strings.TrimSpace(content)
	// Unix or Windows path
	match, _ := regexp.MatchString(`^([A-Za-z]:\\|/)[^\x00]+$`, content)
	return match
}

// isJSON checks if the content is valid JSON
func isJSON(content string) bool {
	content = strings.TrimSpace(content)
	if (strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}")) ||
		(strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]")) {
		// Try to parse as JSON
		var js interface{}
		return json.Unmarshal([]byte(content), &js) == nil
	}
	return false
}
