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
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// ShowAboutDialog displays the about dialog
func ShowAboutDialog(window fyne.Window, app fyne.App) {
	aboutWindow := app.NewWindow("About FyClip")
	aboutWindow.Resize(fyne.NewSize(400, 300))
	
	title := widget.NewLabel("FyClip - Advanced Clipboard Manager")
	title.TextStyle.Bold = true
	
	hyperlink := widget.NewHyperlink("github.com/Sarwarhridoy4", nil)
	if err := hyperlink.SetURLFromString("https://github.com/Sarwarhridoy4"); err != nil {
		fmt.Printf("Warning: Failed to set hyperlink URL: %v\n", err)
	}
	
	content := container.NewVBox(
		title,
		widget.NewLabel(""),
		widget.NewLabel("Developed by: Sarwar Hossain"),
		widget.NewLabel("Email: sarwarhridoy4@gmail.com"),
		hyperlink,
		widget.NewLabel(""),
		widget.NewLabel("Version: 2.0.0"),
		widget.NewLabel("Built with Fyne v2.5+"),
	)
	
	aboutWindow.SetContent(content)
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