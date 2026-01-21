// File: internal/ui/preview.go
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

const maxPreviewImageSize = 5 * 1024 * 1024 // 5MB

// PreviewPane displays preview of selected item
type PreviewPane struct {
	manager   *clipboard.Manager
	text      *widget.RichText
	image     *canvas.Image
	metaText  *canvas.Text
	container *fyne.Container
}

// NewPreviewPane creates a new preview pane
func NewPreviewPane(manager *clipboard.Manager) *PreviewPane {
	pp := &PreviewPane{manager: manager}

	// Rich text preview (selectable + copyable)
	pp.text = widget.NewRichText()
	pp.text.Wrapping = fyne.TextWrapWord

	// Image preview
	pp.image = canvas.NewImageFromResource(theme.BrokenImageIcon())
	pp.image.FillMode = canvas.ImageFillContain
	pp.image.Hide()

	// Metadata overlay
	pp.metaText = canvas.NewText("", theme.ForegroundColor())
	pp.metaText.TextSize = theme.TextSize() - 2
	pp.metaText.Alignment = fyne.TextAlignLeading
	pp.metaText.Hide()

	pp.container = container.NewStack(
		pp.text,
		container.NewBorder(nil, pp.metaText, nil, nil, pp.image),
	)

	pp.setPlaceholder()
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

// ---------------- TEXT ----------------

func (pp *PreviewPane) showText(item clipboard.Item) {
	pp.image.Hide()
	pp.metaText.Hide()
	pp.text.Show()

	pp.text.Segments = syntaxHighlight(item.Content)
	pp.text.Refresh()
}

func (pp *PreviewPane) setPlaceholder() {
	pp.text.Segments = []widget.RichTextSegment{
		&widget.TextSegment{
			Text: "Select an item to preview...",
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNameDisabled,
			},
		},
	}
	pp.text.Refresh()
}

// ---------------- IMAGE ----------------

func (pp *PreviewPane) showImage(item clipboard.Item) {
	pp.text.Hide()

	raw, err := base64.StdEncoding.DecodeString(item.ImageData)
	if err != nil || len(raw) == 0 || len(raw) > maxPreviewImageSize {
		pp.clear()
		return
	}

	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		pp.clear()
		return
	}

	pp.image.Image = img
	pp.image.Show()
	pp.image.Refresh()

	pp.metaText.Text = fmt.Sprintf(
		"%s • %dx%d • %d KB • %s",
		strings.ToUpper(format),
		img.Bounds().Dx(),
		img.Bounds().Dy(),
		len(raw)/1024,
		item.Timestamp.Format("2006-01-02 15:04:05"),
	)
	pp.metaText.Show()
	pp.metaText.Refresh()
}

// ---------------- CLEAR ----------------

func (pp *PreviewPane) clear() {
	pp.image.Hide()
	pp.metaText.Hide()
	pp.text.Show()
	pp.setPlaceholder()
}


func syntaxHighlight(text string) []widget.RichTextSegment {
	lower := strings.ToLower(strings.TrimSpace(text))

	switch {
	case strings.HasPrefix(lower, "{") || strings.HasPrefix(lower, "["):
		return highlightJSON(text)
	case strings.Contains(lower, "package ") || strings.Contains(lower, "func "):
		return highlightCode(text)
	case strings.HasPrefix(lower, "#!") || strings.Contains(lower, "bash"):
		return highlightCode(text)
	default:
		return plainText(text)
	}
}

func plainText(t string) []widget.RichTextSegment {
	return []widget.RichTextSegment{
		&widget.TextSegment{
			Text: t,
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNameForeground,
			},
		},
	}
}

func highlightCode(t string) []widget.RichTextSegment {
	return []widget.RichTextSegment{
		&widget.TextSegment{
			Text: t,
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNamePrimary,
				TextStyle: fyne.TextStyle{Monospace: true},
			},
		},
	}
}

func highlightJSON(t string) []widget.RichTextSegment {
	return []widget.RichTextSegment{
		&widget.TextSegment{
			Text: t,
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNameSuccess,
				TextStyle: fyne.TextStyle{Monospace: true},
			},
		},
	}
}
