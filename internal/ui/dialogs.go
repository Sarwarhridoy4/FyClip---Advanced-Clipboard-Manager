// File: internal/ui/dialogs.go
package ui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg" // Register decoder
	"image/png"
	_ "image/png" // Register decoder
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

const (
	appVersion   = "2.0.0"
	appCopyright = "© 2024-2026"
)

// ShowAboutDialog displays a professional about dialog
func ShowAboutDialog(window fyne.Window, app fyne.App) {
	aboutWindow := app.NewWindow("About FyClip")
	aboutWindow.Resize(fyne.NewSize(450, 550))
	aboutWindow.SetFixedSize(true)

	// App icon
	appIcon := widget.NewIcon(theme.FyneLogo())
	appIcon.Resize(fyne.NewSize(64, 64))

	// Title
	title := widget.NewLabel("FyClip")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle.Bold = true

	// Tagline
	tagline := widget.NewLabel("Advanced Clipboard Manager")
	tagline.Alignment = fyne.TextAlignCenter
	tagline.TextStyle.Italic = true

	// Version info
	versionLabel := widget.NewLabel(fmt.Sprintf("Version %s", appVersion))
	versionLabel.Alignment = fyne.TextAlignCenter

	// Separator
	separator1 := widget.NewSeparator()

	// Description
	descLabel := widget.NewLabel("A modular, high-performance clipboard manager\nbuilt with Go and Fyne.")
	descLabel.Alignment = fyne.TextAlignCenter
	descLabel.Wrapping = fyne.TextWrapWord

	// Features section
	featuresTitle := widget.NewLabel("Features")
	featuresTitle.Alignment = fyne.TextAlignCenter
	featuresTitle.TextStyle.Bold = true

	features := widget.NewLabel("• Clipboard History\n• Rich Text/HTML Support\n• File Path History\n• Encrypted Storage\n• Snippets & Templates\n• Global Hotkey Quick Panel")
	features.Alignment = fyne.TextAlignLeading
	features.Wrapping = fyne.TextWrapWord

	// Separator
	separator2 := widget.NewSeparator()

	// Developer info
	devTitle := widget.NewLabel("Developer")
	devTitle.Alignment = fyne.TextAlignCenter
	devTitle.TextStyle.Bold = true

	devName := widget.NewLabel("Sarwar Hossain")
	devName.Alignment = fyne.TextAlignCenter
	devName.TextStyle.Bold = true

	// Contact links
	githubLink := widget.NewHyperlink("GitHub", nil)
	_ = githubLink.SetURLFromString("https://github.com/Sarwarhridoy4")
	githubLink.Alignment = fyne.TextAlignCenter

	emailLink := widget.NewHyperlink("Email", nil)
	_ = emailLink.SetURLFromString("mailto:sarwarhridoy4@gmail.com")
	emailLink.Alignment = fyne.TextAlignCenter

	contactBox := container.NewHBox(
		layout.NewSpacer(),
		githubLink,
		widget.NewLabel("  |  "),
		emailLink,
		layout.NewSpacer(),
	)

	// Separator
	separator3 := widget.NewSeparator()

	// License
	licenseLabel := widget.NewLabel("MIT License")
	licenseLabel.Alignment = fyne.TextAlignCenter
	licenseLabel.TextStyle.Italic = true

	copyrightLabel := widget.NewLabel(fmt.Sprintf("%s Sarwar Hossain. All rights reserved.", appCopyright))
	copyrightLabel.Alignment = fyne.TextAlignCenter

	// Technology stack
	techTitle := widget.NewLabel("Built with")
	techTitle.Alignment = fyne.TextAlignCenter

	techStack := widget.NewLabel("• Go 1.21+\n• Fyne v2.7+\n• AES-256-GCM Encryption")
	techStack.Alignment = fyne.TextAlignCenter

	// Close button
	closeBtn := widget.NewButton("Close", func() {
		aboutWindow.Close()
	})
	closeBtn.Importance = widget.HighImportance

	// Main content
	content := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), appIcon, layout.NewSpacer()),
		title,
		tagline,
		versionLabel,
		separator1,
		descLabel,
		container.NewPadded(featuresTitle),
		container.NewPadded(features),
		separator2,
		container.NewPadded(devTitle),
		devName,
		contactBox,
		separator3,
		container.NewPadded(techTitle),
		techStack,
		separator3,
		licenseLabel,
		copyrightLabel,
		container.NewPadded(closeBtn),
	)

	// Scroll container for smaller screens
	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(450, 550))

	aboutWindow.SetContent(scrollContent)
	aboutWindow.CenterOnScreen()
	aboutWindow.Show()
}

// ShowNotification shows a temporary notification
func ShowNotification(app fyne.App, message string) {
	if app != nil {
		app.SendNotification(&fyne.Notification{
			Title:   "FyClip",
			Content: message,
		})
	}
}

// ShowPopup displays a temporary popup message
func ShowPopup(window fyne.Window, message string, duration time.Duration) {
	if window == nil || window.Canvas() == nil {
		return
	}

	popupContent := widget.NewCard("", "", widget.NewLabel(message))
	popup := widget.NewPopUp(popupContent, window.Canvas())

	contentPos := window.Content().Position()
	popupPos := contentPos.Add(fyne.NewPos(10, 40))
	popup.Move(popupPos)
	popup.Resize(fyne.NewSize(200, 60))
	popup.Show()

	time.AfterFunc(duration, func() {
		fyne.Do(func() {
			popup.Hide()
		})
	})
}

// ShowImageFormatDialog shows a dialog to select the image format (PNG/JPEG)
func ShowImageFormatDialog(window fyne.Window, callback func(format string, err error)) {
	selectedFormat := "png" // Default to PNG

	radio := widget.NewRadioGroup([]string{"PNG", "JPEG"}, func(s string) {
		if s == "JPEG" {
			selectedFormat = "jpeg"
		} else {
			selectedFormat = "png"
		}
	})
	radio.SetSelected("PNG") // Set initial selection

	dialog.ShowCustomConfirm(
		"Select Image Format",
		"Save",
		"Cancel",
		radio,
		func(confirmed bool) {
			if confirmed {
				callback(selectedFormat, nil)
			} else {
				callback("", fmt.Errorf("format selection cancelled"))
			}
		},
		window,
	)
}

// SaveImage saves an image item to file
func SaveImage(item clipboard.Item, filename, format string) error {
	if item.Type != clipboard.TypeImage || item.ImageData == "" {
		return fmt.Errorf("no image data available")
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(item.ImageData)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image format: %w", err)
	}

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Encode in requested format
	switch format {
	case "png":
		if err := png.Encode(file, img); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	case "jpeg", "jpg":
		if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 95}); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// ShowFeaturesDialog displays a dialog explaining how each feature works
func ShowFeaturesDialog(window fyne.Window, app fyne.App) {
	featuresWindow := app.NewWindow("Features Guide")
	featuresWindow.Resize(fyne.NewSize(500, 600))
	featuresWindow.SetFixedSize(true)

	// Title
	title := widget.NewLabel("FyClip Features Guide")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle.Bold = true

	// Separator
	separator := widget.NewSeparator()

	// Features list with descriptions
	features := []struct {
		name        string
		description string
	}{
		{
			name:        "Clipboard History",
			description: "Automatically captures and stores all clipboard content (text, images, HTML, files). Access your history anytime with the global hotkey.",
		},
		{
			name:        "Pin/Unpin Items",
			description: "Pin important items to keep them at the top of your history. Pinned items are never automatically cleared.",
		},
		{
			name:        "Favorites Filter",
			description: "Toggle between showing all items or only pinned favorites. Quickly access your most important clipboard entries.",
		},
		{
			name:        "Pause Monitoring",
			description: "Temporarily pause clipboard monitoring for 5 minutes. Useful when copying sensitive information you don't want stored.",
		},
		{
			name:        "Search",
			description: "Search through your clipboard history by content. Quickly find specific items you copied earlier.",
		},
		{
			name:        "Preview Pane",
			description: "View detailed preview of selected items. Supports text, images, HTML, and file information with syntax highlighting.",
		},
		{
			name:        "Export",
			description: "Export selected items to files. Text items are saved as .txt files, images can be saved as PNG or JPEG.",
		},
		{
			name:        "Snippets",
			description: "Create and manage text snippets for quick insertion. Use abbreviations for faster access.\n\nTemplate Variables:\n• {{date}} - Current date (YYYY-MM-DD)\n• {{time}} - Current time (HH:MM:SS)\n• {{datetime}} - Full date and time\n• {{year}} - Current year\n• {{month}} - Current month (01-12)\n• {{day}} - Current day (01-31)\n• {{clipboard}} - Current clipboard content\n\nTo use: Click Snippets button in toolbar, select a snippet, and click Use to copy to clipboard.",
		},
		{
			name:        "Quick Panel",
			description: "Access clipboard history with a global hotkey (Ctrl+Shift+V). Paste items without switching windows.",
		},
		{
			name:        "Clear History",
			description: "Clear all unpinned items from history. Pinned items are preserved for future use.",
		},
		{
			name:        "Refresh",
			description: "Manually refresh the clipboard history list. Useful if items aren't updating automatically.",
		},
	}

	// Create feature items
	featureItems := []fyne.CanvasObject{}
	for _, f := range features {
		nameLabel := widget.NewLabel(f.name)
		nameLabel.TextStyle.Bold = true
		nameLabel.Wrapping = fyne.TextWrapWord

		descLabel := widget.NewLabel(f.description)
		descLabel.Wrapping = fyne.TextWrapWord

		featureBox := container.NewVBox(
			nameLabel,
			descLabel,
			widget.NewSeparator(),
		)
		featureItems = append(featureItems, featureBox)
	}

	// Close button
	closeBtn := widget.NewButton("Close", func() {
		featuresWindow.Close()
	})
	closeBtn.Importance = widget.HighImportance

	// Main content
	content := container.NewVBox(
		title,
		separator,
	)
	for _, item := range featureItems {
		content.Add(item)
	}
	content.Add(container.NewPadded(closeBtn))

	// Scroll container for smaller screens
	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(500, 600))

	featuresWindow.SetContent(scrollContent)
	featuresWindow.CenterOnScreen()
	featuresWindow.Show()
}
