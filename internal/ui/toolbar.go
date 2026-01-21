// File: internal/ui/toolbar.go
package ui

import (
	"fmt"
	"path/filepath"
	"strings"

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
	
	go func() {
		if t.manager.TogglePin(index) {
			item, _ := t.manager.GetItem(index)
			status := "Pinned"
			if !item.Pinned {
				status = "Unpinned"
			}
			
			t.manager.Shutdown() // Force save
			ShowNotification(t.app, status+" item!")
		}
	}()
}

// onDelete handles delete button
func (t *Toolbar) onDelete() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}
	
	go func() {
		if err := t.manager.Delete(index); err != nil {
			ShowNotification(t.app, err.Error())
			return
		}
		
		t.manager.Shutdown() // Force save
		t.list.UnselectAll()
		ShowNotification(t.app, "Deleted item!")
	}()
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
			
			go func() {
				t.manager.ClearUnpinned()
				t.manager.Shutdown() // Force save
				t.list.UnselectAll()
				ShowNotification(t.app, "Cleared history!")
			}()
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
	
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		
		filename := writer.URI().Path()
		ext := strings.ToLower(filepath.Ext(filename))
		
		format := "png"
		if ext == ".jpg" || ext == ".jpeg" {
			format = "jpeg"
		} else if ext != ".png" {
			filename += ".png"
		}
		
		go func() {
			if err := SaveImage(item, filename, format); err != nil {
				ShowNotification(t.app, fmt.Sprintf("Save failed: %v", err))
				return
			}
			ShowNotification(t.app, fmt.Sprintf("Saved as %s", filepath.Base(filename)))
		}()
	}, t.window)
}