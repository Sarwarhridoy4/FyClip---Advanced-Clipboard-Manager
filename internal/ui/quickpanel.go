package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

type QuickPanel struct {
	manager    *clipboard.Manager
	window     fyne.Window
	popup      *widget.PopUp
	visible    bool
	onSelect   func(item clipboard.Item)
	items      []clipboard.Item
	list       *widget.List
	selectedID widget.ListItemID
}

func NewQuickPanel(manager *clipboard.Manager, window fyne.Window, onSelect func(item clipboard.Item)) *QuickPanel {
	return &QuickPanel{
		manager: manager,
		window:  window,
		onSelect: onSelect,
	}
}

func (q *QuickPanel) Show() {
	if q.visible {
		return
	}

	q.items = q.manager.GetFiltered()
	if len(q.items) == 0 {
		return
	}

	maxItems := 15
	if len(q.items) > maxItems {
		q.items = q.items[:maxItems]
	}

	q.selectedID = 0
	q.list = widget.NewList(
		func() int { return len(q.items) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				canvas.NewRectangle(nil),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(q.items) {
				return
			}
			itm := q.items[id]
			row := item.(*fyne.Container)
			label := row.Objects[0].(*widget.Label)
			bg := row.Objects[1].(*canvas.Rectangle)
			text := fmt.Sprintf("%d. %s", id+1, q.truncate(itm.DisplayText(80)))
			if itm.Pinned {
				text = "★ " + text
			}
			label.SetText(text)
			if id == q.selectedID {
				bg.FillColor = theme.Color(theme.ColorNamePrimary)
				bg.Show()
			} else {
				bg.Hide()
			}
		},
	)

	q.list.OnSelected = func(id widget.ListItemID) {
		q.selectItem(id)
		q.Hide()
	}

	instructions := widget.NewLabel("↑↓ navigate • Enter select • Esc close")
	instructions.Alignment = fyne.TextAlignCenter
	instructions.TextStyle.Italic = true

	content := container.NewBorder(nil, instructions, nil, nil, q.list)

	width := float32(420)
	height := float32(320)
	if len(q.items) < 5 {
		height = float32(len(q.items)*40 + 60)
	}

	popup := widget.NewPopUp(content, q.window.Canvas())
	popup.Resize(fyne.NewSize(width, height))

	canvasSize := q.window.Canvas().Size()
	x := (canvasSize.Width - width) / 2
	y := float32(40)
	popup.ShowAtPosition(fyne.NewPos(x, y))

	q.popup = popup
	q.visible = true

	q.window.Canvas().Focus(q.list)
}

func (q *QuickPanel) Hide() {
	if !q.visible {
		return
	}
	q.visible = false
	if q.popup != nil {
		q.popup.Hide()
		q.popup = nil
	}
}

func (q *QuickPanel) Toggle() {
	if q.visible {
		q.Hide()
	} else {
		q.Show()
	}
}

func (q *QuickPanel) IsVisible() bool {
	return q.visible
}

func (q *QuickPanel) selectItem(index int) {
	if index < 0 || index >= len(q.items) {
		return
	}
	item := q.items[index]
	if q.onSelect != nil {
		q.onSelect(item)
	}
}

func (q *QuickPanel) truncate(s string) string {
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}

func (q *QuickPanel) Navigate(delta int) {
	if !q.visible || q.list == nil || len(q.items) == 0 {
		return
	}
	newIndex := int(q.selectedID) + delta
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(q.items) {
		newIndex = len(q.items) - 1
	}
	q.selectedID = widget.ListItemID(newIndex)
	q.list.Select(q.selectedID)
	q.list.Refresh()
}

func (q *QuickPanel) SelectCurrent() {
	if !q.visible || q.list == nil {
		return
	}
	q.selectItem(int(q.selectedID))
	q.Hide()
}

func (q *QuickPanel) HandleKeyEvent(key *fyne.KeyEvent) bool {
	if !q.visible {
		return false
	}
	switch key.Name {
	case fyne.KeyDown:
		q.Navigate(1)
		return true
	case fyne.KeyUp:
		q.Navigate(-1)
		return true
	case fyne.KeyEnter, fyne.KeyTab:
		q.SelectCurrent()
		return true
	case fyne.KeyEscape:
		q.Hide()
		return true
	}
	return false
}
