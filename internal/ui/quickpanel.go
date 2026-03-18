// File: internal/ui/quickpanel.go
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// QuickPanel manages the quick paste overlay
type QuickPanel struct {
	manager    *clipboard.Manager
	window     fyne.Window
	dialog     *dialog.CustomDialog
	visible    bool
	onSelect   func(item clipboard.Item)
	keyTypedID fyne.CanvasObject
}

// NewQuickPanel creates a new quick panel
func NewQuickPanel(manager *clipboard.Manager, window fyne.Window, onSelect func(item clipboard.Item)) *QuickPanel {
	q := &QuickPanel{
		manager:  manager,
		window:   window,
		onSelect: onSelect,
	}
	
	return q
}

// Show displays the quick panel
func (q *QuickPanel) Show() {
	if q.visible {
		return
	}
	
	items := q.manager.GetFiltered()
	if len(items) > 9 {
		items = items[:9]
	}

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(items) {
				itm := items[id]
				item.(*widget.Label).SetText(fmt.Sprintf("%d. %s", id+1, q.truncate(itm.DisplayText(50))))
			}
		},
	)
	
	list.OnSelected = func(id widget.ListItemID) {
		q.selectItem(id, items)
		q.Hide()
	}
	
	instructions := widget.NewLabel("↑↓ to navigate • Enter to select • Esc to close")
	instructions.Alignment = fyne.TextAlignCenter
	
	content := container.NewVBox(
		container.NewMax(list),
		widget.NewSeparator(),
		instructions,
	)
	
	q.dialog = dialog.NewCustomWithoutButtons("Quick Paste", content, q.window)
	q.dialog.SetOnClosed(func() {
		q.visible = false
	})
	
	q.visible = true
	q.dialog.Show()
}

// Hide hides the quick panel
func (q *QuickPanel) Hide() {
	if !q.visible {
		return
	}
	
	q.visible = false
	if q.dialog != nil {
		q.dialog.Hide()
	}
}

// Toggle shows or hides the quick panel
func (q *QuickPanel) Toggle() {
	if q.visible {
		q.Hide()
	} else {
		q.Show()
	}
}

// IsVisible returns whether the panel is visible
func (q *QuickPanel) IsVisible() bool {
	return q.visible
}

// selectItem selects an item at the given index
func (q *QuickPanel) selectItem(index int, items []clipboard.Item) {
	if index < 0 || index >= len(items) {
		return
	}
	
	item := items[index]
	
	if q.onSelect != nil {
		q.onSelect(item)
	}
}

// truncate truncates text to max length
func (q *QuickPanel) truncate(s string) string {
	if len(s) > 50 {
		return s[:47] + "..."
	}
	return s
}
