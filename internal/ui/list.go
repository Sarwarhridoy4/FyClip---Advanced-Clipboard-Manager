// File: internal/ui/list.go
package ui

import (
	"fmt"

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
	app        fyne.App
	window     fyne.Window

	// Multi-select support
	selectionMode bool
	selectedIDs   map[string]bool
}

// NewHistoryList creates a new history list
func NewHistoryList(manager *clipboard.Manager, onSelected func(int), app fyne.App, window fyne.Window) *HistoryList {
	hl := &HistoryList{
		manager:     manager,
		onSelected:  onSelected,
		app:         app,
		window:      window,
		selectionMode: false,
		selectedIDs: make(map[string]bool),
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
		if hl.selectionMode {
			// In selection mode, toggle selection instead of selecting
			hl.ToggleSelection(id)
		} else if hl.onSelected != nil {
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

// SetSelectionMode enables or disables multi-select mode
func (hl *HistoryList) SetSelectionMode(enabled bool) {
	hl.selectionMode = enabled
	if !enabled {
		hl.ClearSelection()
	}
	if hl.list != nil {
		hl.list.Refresh()
	}
}

// IsSelectionMode returns whether multi-select mode is enabled
func (hl *HistoryList) IsSelectionMode() bool {
	return hl.selectionMode
}

// ToggleSelection toggles the selection state of an item
func (hl *HistoryList) ToggleSelection(index int) {
	item, ok := hl.manager.GetItem(index)
	if !ok {
		return
	}

	if hl.selectedIDs[item.ID] {
		delete(hl.selectedIDs, item.ID)
	} else {
		hl.selectedIDs[item.ID] = true
	}

	if hl.list != nil {
		hl.list.Refresh()
	}
}

// GetSelectedIDs returns all selected item IDs
func (hl *HistoryList) GetSelectedIDs() []string {
	ids := make([]string, 0, len(hl.selectedIDs))
	for id := range hl.selectedIDs {
		ids = append(ids, id)
	}
	return ids
}

// GetSelectedCount returns the number of selected items
func (hl *HistoryList) GetSelectedCount() int {
	return len(hl.selectedIDs)
}

// ClearSelection clears all selected items
func (hl *HistoryList) ClearSelection() {
	hl.selectedIDs = make(map[string]bool)
	if hl.list != nil {
		hl.list.Refresh()
	}
}

// SelectAll selects all visible items
func (hl *HistoryList) SelectAll() {
	count := hl.manager.GetFilteredCount()
	for i := 0; i < count; i++ {
		item, ok := hl.manager.GetItem(i)
		if ok {
			hl.selectedIDs[item.ID] = true
		}
	}
	if hl.list != nil {
		hl.list.Refresh()
	}
}

// createTemplate creates the list item template
func (hl *HistoryList) createTemplate() fyne.CanvasObject {
	pinBtn := widget.NewButtonWithIcon("", theme.RadioButtonIcon(), nil)
	pinBtn.Importance = widget.LowImportance

	// Checkbox for multi-select mode
	checkBox := widget.NewCheck("", nil)
	checkBox.Hide() // Hidden by default

	typeIcon := widget.NewIcon(theme.FileTextIcon())
	contentLabel := widget.NewLabel("Template")
	timeLabel := widget.NewLabel("00:00")
	timeLabel.TextStyle.Monospace = true

	// Layout: checkbox (hidden by default), pin button, type icon, content, time
	content := container.NewHBox(
		checkBox,
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
	// Now we have 6 objects: checkbox, pin button, type icon, content, spacer, time
	if len(innerContainer.Objects) < 6 {
		return
	}

	// Update checkbox (index 0) - only in selection mode
	checkBox := innerContainer.Objects[0].(*widget.Check)
	if hl.selectionMode {
		checkBox.Show()
		checkBox.SetChecked(hl.selectedIDs[item.ID])
		// Update checkbox callback to toggle selection
		currentIndex := index
		checkBox.OnChanged = func(checked bool) {
			// Get current item ID at this index
			item, ok := hl.manager.GetItem(currentIndex)
			if ok {
				if checked {
					hl.selectedIDs[item.ID] = true
				} else {
					delete(hl.selectedIDs, item.ID)
				}
			}
			hl.Refresh()
		}
	} else {
		checkBox.Hide()
	}

	// Update pin button (index 1)
	pinBtn := innerContainer.Objects[1].(*widget.Button)
	if item.Pinned {
		pinBtn.SetIcon(theme.ConfirmIcon())
	} else {
		pinBtn.SetIcon(theme.RadioButtonIcon())
	}

	// In selection mode, hide the pin button to make room for checkbox
	if hl.selectionMode {
		pinBtn.Hide()
	} else {
		pinBtn.Show()
	}

	// Capture item ID to avoid stale index issue
	itemID := item.ID
	pinBtn.OnTapped = func() {
		// Find current index of this item
		currentIndex := hl.manager.FindIndexByID(itemID)
		if currentIndex >= 0 {
			if hl.manager.TogglePin(currentIndex) {
				hl.manager.SaveHistory()
				hl.Refresh()
			}
		}
	}

	// Update type icon (index 2)
	typeIcon := innerContainer.Objects[2].(*widget.Icon)
	switch item.Type {
	case clipboard.TypeText:
		typeIcon.SetResource(preferredDocIcon())
	case clipboard.TypeImage:
		typeIcon.SetResource(theme.FileImageIcon())
	case clipboard.TypeHTML:
		typeIcon.SetResource(theme.FileTextIcon())
	case clipboard.TypeFile:
		typeIcon.SetResource(theme.FolderOpenIcon())
	default:
		typeIcon.SetResource(theme.FileIcon())
	}
	typeIcon.Refresh()

	// Update content label (index 3)
	contentLabel := innerContainer.Objects[3].(*widget.Label)
	contentLabel.SetText(item.DisplayText(80))

	// Update time label (index 5)
	timeLabel := innerContainer.Objects[5].(*widget.Label)
	meta := item.TimeAgo()
	if item.CopyCount > 0 {
		meta = fmt.Sprintf("%s x%d", meta, item.CopyCount)
	}
	timeLabel.SetText(meta)
}

func preferredDocIcon() fyne.Resource {
	if res := theme.FileTextIcon(); res != nil {
		return res
	}
	return theme.DocumentIcon()
}
