// File: internal/ui/list.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// HistoryList displays the clipboard history
type HistoryList struct {
	manager    *clipboard.Manager
	list       *widget.List
	onSelected func(int)
}

// NewHistoryList creates a new history list
func NewHistoryList(manager *clipboard.Manager, onSelected func(int)) *HistoryList {
	hl := &HistoryList{
		manager:    manager,
		onSelected: onSelected,
	}
	
	hl.list = widget.NewList(
		func() int {
			return hl.manager.GetFilteredCount()
		},
		func() fyne.CanvasObject {
			return hl.createTemplate()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			hl.updateItem(id, obj)
		},
	)
	
	hl.list.OnSelected = func(id widget.ListItemID) {
		if hl.onSelected != nil {
			hl.onSelected(id)
		}
	}
	
	return hl
}

// Build returns the list widget
func (hl *HistoryList) Build() fyne.CanvasObject {
	return hl.list
}

// Refresh updates the list
func (hl *HistoryList) Refresh() {
	if hl.list != nil {
		hl.list.Refresh()
	}
}

// UnselectAll clears selection
func (hl *HistoryList) UnselectAll() {
	if hl.list != nil {
		hl.list.UnselectAll()
	}
}

// createTemplate creates the list item template
func (hl *HistoryList) createTemplate() fyne.CanvasObject {
	pinBtn := widget.NewButtonWithIcon("", theme.RadioButtonIcon(), nil)
	pinBtn.Importance = widget.LowImportance
	
	typeIcon := widget.NewIcon(theme.FileTextIcon())
	contentLabel := widget.NewLabel("Template")
	timeLabel := widget.NewLabel("00:00")
	timeLabel.TextStyle.Monospace = true
	
	content := container.NewHBox(
		pinBtn,
		typeIcon,
		contentLabel,
		widget.NewLabel(""), // spacer
		timeLabel,
	)
	
	return container.NewPadded(content)
}

// updateItem updates a list item with data
func (hl *HistoryList) updateItem(index int, obj fyne.CanvasObject) {
	item, ok := hl.manager.GetItem(index)
	if !ok {
		return
	}
	
	container := obj.(*fyne.Container)
	if len(container.Objects) == 0 {
		return
	}
	
	innerContainer := container.Objects[0].(*fyne.Container)
	if len(innerContainer.Objects) < 5 {
		return
	}
	
	// Update pin button
	pinBtn := innerContainer.Objects[0].(*widget.Button)
	if item.Pinned {
		pinBtn.SetIcon(theme.ConfirmIcon())
	} else {
		pinBtn.SetIcon(theme.RadioButtonIcon())
	}
	pinBtn.OnTapped = func() {
		go func() {
			if hl.manager.TogglePin(index) {
				hl.manager.Shutdown() // Force save
				hl.Refresh()
			}
		}()
	}
	
	// Update type icon
	typeIcon := innerContainer.Objects[1].(*widget.Icon)
	if item.Type == clipboard.TypeText {
		typeIcon.SetResource(preferredDocIcon())
	} else {
		typeIcon.SetResource(theme.FileImageIcon())
	}
	typeIcon.Refresh()
	
	// Update content label
	contentLabel := innerContainer.Objects[2].(*widget.Label)
	contentLabel.SetText(item.DisplayText(80))
	
	// Update time label
	timeLabel := innerContainer.Objects[4].(*widget.Label)
	timeLabel.SetText(item.Timestamp.Format("15:04"))
}

func preferredDocIcon() fyne.Resource {
	if res := theme.FileTextIcon(); res != nil {
		return res
	}
	return theme.DocumentIcon()
}
