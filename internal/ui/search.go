// File: internal/ui/search.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// SearchBar provides search functionality
type SearchBar struct {
	manager  *clipboard.Manager
	list     *HistoryList
	entry    *widget.Entry
	clearBtn *widget.Button
	root     *fyne.Container
}

// NewSearchBar creates a new search bar
func NewSearchBar(manager *clipboard.Manager, list *HistoryList) *SearchBar {
	sb := &SearchBar{
		manager: manager,
		list:    list,
	}

	sb.entry = widget.NewEntry()
	sb.entry.SetPlaceHolder("Search clipboard history...")
	sb.entry.OnChanged = sb.onSearch
	sb.clearBtn = widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		sb.entry.SetText("")
		if sb.list != nil {
			sb.list.UnselectAll()
		}
	})
	sb.clearBtn.Importance = widget.LowImportance
	sb.root = container.NewBorder(nil, nil, nil, sb.clearBtn, sb.entry)

	return sb
}

// Build returns the search widget
func (sb *SearchBar) Build() fyne.CanvasObject {
	return sb.root
}

// onSearch handles search query changes
func (sb *SearchBar) onSearch(query string) {
	sb.manager.SetSearch(query)
	if sb.list != nil {
		sb.list.UnselectAll()
	}
}

// Focus sets focus to the search entry
func (sb *SearchBar) Focus(window fyne.Window) {
	if sb.entry != nil && window != nil {
		window.Canvas().Focus(sb.entry)
	}
}

// Clear clears the search input
func (sb *SearchBar) Clear() {
	if sb.entry != nil {
		sb.entry.SetText("")
	}
}
