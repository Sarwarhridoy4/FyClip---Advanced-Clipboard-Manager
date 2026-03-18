package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// PreviewPane displays preview of selected item
type PreviewPane struct {
	manager   *clipboard.Manager
	text      *widget.RichText
	scroll    *container.Scroll
	image     *canvas.Image
	copyBtn   *widget.Button
	container *fyne.Container

	rawText string

	lastItemID   string
	lastItemType clipboard.ItemType
	hasSelection bool
}

// NewPreviewPane creates a new preview pane
func NewPreviewPane(manager *clipboard.Manager) *PreviewPane {
	pp := &PreviewPane{
		manager: manager,
	}

	// Markdown-capable rich text
	pp.text = pp.newMarkdownText("_Select an item to preview..._")

	pp.scroll = container.NewVScroll(pp.text)
	pp.scroll.Hide()

	pp.image = canvas.NewImageFromResource(theme.BrokenImageIcon())
	pp.image.FillMode = canvas.ImageFillContain
	pp.image.Hide()

	pp.copyBtn = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		if pp.rawText == "" {
			return
		}
		fyne.CurrentApp().Clipboard().SetContent(pp.rawText)
	})
	pp.copyBtn.Importance = widget.LowImportance
	pp.copyBtn.Hide()

	pp.container = container.NewBorder(
		container.NewHBox(layoutSpacer(), pp.copyBtn),
		nil,
		nil,
		nil,
		container.NewStack(pp.scroll, pp.image),
	)

	pp.showPlaceholder()
	return pp
}

// Build returns the preview widget
func (pp *PreviewPane) Build() fyne.CanvasObject {
	return pp.container
}

// Refresh updates the preview
func (pp *PreviewPane) Refresh() {
	item, ok := pp.manager.GetSelected()
	if !ok {
		if !pp.hasSelection {
			return
		}
		pp.clear()
		return
	}

	if pp.hasSelection && item.ID == pp.lastItemID && item.Type == pp.lastItemType {
		return
	}

	switch item.Type {
	case clipboard.TypeText:
		pp.showText(item)
	case clipboard.TypeImage:
		pp.showImage(item)
	case clipboard.TypeHTML:
		pp.showHTML(item)
	case clipboard.TypeFile:
		pp.showFile(item)
	}
}

// showText renders markdown text with scrollbar
func (pp *PreviewPane) showText(item clipboard.Item) {
	pp.image.Hide()
	pp.copyBtn.Show()

	pp.rawText = item.Content

	// Check if content is JSON and pretty-print it
	content := item.Content
	if isJSON(content) {
		if prettyJSON, err := json.MarshalIndent(json.RawMessage(content), "", "  "); err == nil {
			content = string(prettyJSON)
			content = "**JSON (Pretty)**\n\n" + content
		}
	}

	pp.setMarkdown(content)

	pp.scroll.Show()
	pp.scroll.ScrollToTop()
	pp.markRendered(item)
}

// showHTML renders HTML content
func (pp *PreviewPane) showHTML(item clipboard.Item) {
	pp.image.Hide()
	pp.copyBtn.Show()

	pp.rawText = item.Content

	// Show plain text version of HTML
	content := item.GetDisplayContent()
	if item.HTMLContent != "" {
		content = "**HTML Content**\n\n" + content + "\n\n---\n\n*HTML available - copy as HTML to preserve formatting*"
	}

	pp.setMarkdown(content)

	pp.scroll.Show()
	pp.scroll.ScrollToTop()
	pp.markRendered(item)
}

// showFile displays file information
func (pp *PreviewPane) showFile(item clipboard.Item) {
	pp.image.Hide()
	pp.copyBtn.Show()

	if item.FileInfo == nil {
		pp.setMarkdown("**File**\n\nNo file information available")
		pp.scroll.Show()
		pp.markRendered(item)
		return
	}

	fi := item.FileInfo
	fileType := "File"
	if fi.IsDirectory {
		fileType = "Directory"
	}

	content := fmt.Sprintf("**%s: %s**\n\n", fileType, fi.Name)
	content += fmt.Sprintf("Path: `%s`\n", fi.Path)
	content += fmt.Sprintf("Size: %s\n", formatFileSize(fi.Size))
	content += fmt.Sprintf("Modified: %s\n", fi.ModTime.Format("2006-01-02 15:04:05"))
	content += "\n---\n\n*Copy to clipboard: copies file path*"

	pp.rawText = fi.Path
	pp.setMarkdown(content)

	pp.scroll.Show()
	pp.scroll.ScrollToTop()
	pp.markRendered(item)
}

// isJSON checks if a string is valid JSON
func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") && !strings.HasPrefix(s, "[") {
		return false
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
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

// showImage displays image preview
func (pp *PreviewPane) showImage(item clipboard.Item) {
	pp.copyBtn.Hide()
	pp.rawText = ""

	if item.ImageData == "" {
		pp.clear()
		return
	}

	imageBytes, err := base64.StdEncoding.DecodeString(item.ImageData)
	if err != nil {
		pp.showImageError(item, "Failed to decode image data")
		return
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		pp.showImageError(item, "Invalid image format")
		return
	}

	size := item.Size()
	info := fmt.Sprintf(
		"**%s Image**\n\nCopied: %s\nSize: ~%d bytes\nFormat: %s",
		strings.ToUpper(format),
		item.Timestamp.Format("2006-01-02 15:04:05"),
		size,
		format,
	)

	pp.setMarkdown(info)
	pp.scroll.Show()
	pp.scroll.ScrollToTop()

	pp.image.Resource = fyne.NewStaticResource("preview", imageBytes)
	pp.image.Show()
	pp.image.Refresh()
	pp.markRendered(item)
}

// showImageError displays an error for image preview
func (pp *PreviewPane) showImageError(item clipboard.Item, errMsg string) {
	pp.copyBtn.Hide()
	pp.rawText = ""

	info := fmt.Sprintf(
		"**Image Error**\n\nCopied: %s\nSize: ~%d bytes\n\n❌ %s",
		item.Timestamp.Format("2006-01-02 15:04:05"),
		item.Size(),
		errMsg,
	)

	pp.setMarkdown(info)
	pp.scroll.Show()
	pp.scroll.ScrollToTop()
	pp.image.Hide()
	pp.markRendered(item)
}

// clear clears the preview
func (pp *PreviewPane) clear() {
	pp.rawText = ""
	pp.copyBtn.Hide()
	pp.hasSelection = false
	pp.lastItemID = ""
	pp.showPlaceholder()
	pp.image.Hide()
}

func (pp *PreviewPane) showPlaceholder() {
	pp.setMarkdown("_Select an item to preview..._")
	pp.scroll.Show()
	pp.scroll.ScrollToTop()
}

func (pp *PreviewPane) setMarkdown(markdown string) {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	pp.text.ParseMarkdown(normalized)
	pp.text.Refresh()
}

func (pp *PreviewPane) newMarkdownText(markdown string) *widget.RichText {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	rt := widget.NewRichTextFromMarkdown(normalized)
	rt.Wrapping = fyne.TextWrapWord
	return rt
}

func (pp *PreviewPane) markRendered(item clipboard.Item) {
	pp.lastItemID = item.ID
	pp.lastItemType = item.Type
	pp.hasSelection = true
}

func layoutSpacer() fyne.CanvasObject {
	return container.NewHBox()
}
