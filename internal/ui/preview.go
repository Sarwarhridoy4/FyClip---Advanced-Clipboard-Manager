package ui

import (
	"bytes"
	"encoding/base64"
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
}

// NewPreviewPane creates a new preview pane
func NewPreviewPane(manager *clipboard.Manager) *PreviewPane {
	pp := &PreviewPane{
		manager: manager,
	}

	// Markdown-capable rich text
	pp.text = widget.NewRichText()
	pp.text.Wrapping = fyne.TextWrapWord

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
		pp.clear()
		return
	}

	switch item.Type {
	case clipboard.TypeText:
		pp.showText(item)
	case clipboard.TypeImage:
		pp.showImage(item)
	}
}

// showText renders markdown text with scrollbar
func (pp *PreviewPane) showText(item clipboard.Item) {
	pp.image.Hide()
	pp.copyBtn.Show()

	pp.rawText = item.Content
	pp.text.ParseMarkdown(item.Content)

	pp.scroll.Show()
	pp.scroll.ScrollToTop()
	pp.text.Refresh()
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

	pp.text.ParseMarkdown(info)
	pp.scroll.Show()
	pp.scroll.ScrollToTop()

	pp.image.Resource = fyne.NewStaticResource("preview", imageBytes)
	pp.image.Show()
	pp.image.Refresh()
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

	pp.text.ParseMarkdown(info)
	pp.scroll.Show()
	pp.scroll.ScrollToTop()
	pp.image.Hide()
}

// clear clears the preview
func (pp *PreviewPane) clear() {
	pp.rawText = ""
	pp.copyBtn.Hide()
	pp.showPlaceholder()
	pp.image.Hide()
}

func (pp *PreviewPane) showPlaceholder() {
	pp.text.ParseMarkdown("_Select an item to preview..._")
	pp.scroll.Show()
	pp.scroll.ScrollToTop()
}

func layoutSpacer() fyne.CanvasObject {
	return container.NewHBox()
}
