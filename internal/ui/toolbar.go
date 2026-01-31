// File: internal/ui/toolbar.go
package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time" // Added import

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// Toolbar provides action buttons
type Toolbar struct {
	window  fyne.Window
	app     fyne.App
	manager *clipboard.Manager
	list    *HistoryList
	
	container *fyne.Container
}

// NewToolbar creates a new toolbar
func NewToolbar(window fyne.Window, app fyne.App, manager *clipboard.Manager, list *HistoryList) *Toolbar {
	return &Toolbar{
		window:  window,
		app:     app,
		manager: manager,
		list:    list,
	}
}

// Build creates the toolbar widget
func (t *Toolbar) Build() fyne.CanvasObject {
	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), t.onCopy)
	pinBtn := widget.NewButtonWithIcon("Pin/Unpin", theme.ViewRefreshIcon(), t.onPin)
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), t.onDelete)
	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), t.onClear)
	saveBtn := widget.NewButtonWithIcon("Save Image", theme.DocumentSaveIcon(), t.onSaveImage)
	
	t.container = container.NewHBox(
		copyBtn,
		pinBtn,
		deleteBtn,
		clearBtn,
		saveBtn,
	)
	
	return t.container
}

// onCopy handles copy button
func (t *Toolbar) onCopy() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}
	
	if err := t.manager.CopyToClipboard(index); err != nil {
		ShowNotification(t.app, fmt.Sprintf("Copy failed: %v", err))
		return
	}
	
	ShowNotification(t.app, "Copied to clipboard!")
}

// onPin handles pin button
func (t *Toolbar) onPin() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}
	
	// Get current state before toggling
	item, ok := t.manager.GetItem(index)
	if !ok {
		ShowNotification(t.app, "Failed to get item!")
		return
	}
	
	currentlyPinned := item.Pinned
	
	// Toggle the pin
	if t.manager.TogglePin(index) {
		t.manager.SaveHistory()
		
		// Show appropriate message
		if currentlyPinned {
			ShowNotification(t.app, "Item unpinned!")
		} else {
			ShowNotification(t.app, "Item pinned!")
		}
		
		// Force list refresh
		if t.list != nil {
			t.list.Refresh()
		}
	} else {
		ShowNotification(t.app, "Failed to toggle pin!")
	}
}

// onDelete handles delete button
func (t *Toolbar) onDelete() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}
	
	if err := t.manager.Delete(index); err != nil {
		ShowNotification(t.app, err.Error())
		return
	}
	
	t.manager.SaveHistory()
	
	// Unselect and refresh
	if t.list != nil {
		t.list.UnselectAll()
		t.list.Refresh()
	}
	
	ShowNotification(t.app, "Item deleted!")
}

// onClear handles clear button
func (t *Toolbar) onClear() {
	dialog.ShowConfirm(
		"Clear History",
		"Are you sure you want to clear all unpinned items?",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			
			t.manager.ClearUnpinned()
			t.manager.SaveHistory()
			
			// Unselect and refresh
			if t.list != nil {
				t.list.UnselectAll()
				t.list.Refresh()
			}
			
			ShowNotification(t.app, "History cleared!")
		},
		t.window,
	)
}

// onSaveImage handles save image button
func (t *Toolbar) onSaveImage() {
	item, ok := t.manager.GetSelected()
	if !ok || item.Type != clipboard.TypeImage {
		ShowNotification(t.app, "Please select an image!")
		return
	}

	// First, ask the user for the desired format
	ShowImageFormatDialog(t.window, func(selectedFormat string, err error) {
		if err != nil { // User cancelled format selection
			return
		}

		// Determine the default extension based on the selected format
		var defaultExtension string
		if selectedFormat == "jpeg" {
			defaultExtension = ".jpeg"
		} else {
			defaultExtension = ".png"
		}

		// Generate a suggested filename with the default extension
		suggestedFilename := fmt.Sprintf("image_%s%s", time.Now().Format("20060102_150405"), defaultExtension)
		
		// Create a file save dialog
		fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()

			filename := writer.URI().Path()
			// Ensure the filename has the correct extension based on user's initial choice
			// or if they changed it in the file dialog
			if !strings.HasSuffix(strings.ToLower(filename), ".png") &&
				!strings.HasSuffix(strings.ToLower(filename), ".jpg") &&
				!strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
				filename += defaultExtension // Append default if no known extension was provided
			}

			// The actual format to save will be based on the selectedFormat,
			// or re-evaluated from the final filename if user changed it.
			formatToSave := selectedFormat
			ext := strings.ToLower(filepath.Ext(filename))
			if ext == ".jpg" || ext == ".jpeg" {
				formatToSave = "jpeg"
			} else if ext == ".png" {
				formatToSave = "png"
			}


			if err := SaveImage(item, filename, formatToSave); err != nil {
				ShowNotification(t.app, fmt.Sprintf("Save failed: %v", err))
				return
			}
			ShowNotification(t.app, fmt.Sprintf("Saved as %s", filepath.Base(filename)))
		}, t.window)
		
		fileSaveDialog.SetFileName(suggestedFilename) // Set the suggested filename
		fileSaveDialog.Show()
	})
}