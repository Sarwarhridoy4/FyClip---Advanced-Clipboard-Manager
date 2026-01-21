// File: internal/ui/search.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// SearchBar provides search functionality
type SearchBar struct {
	manager *clipboard.Manager
	list    *HistoryList
	entry   *widget.Entry
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
	
	return sb
}

// Build returns the search widget
func (sb *SearchBar) Build() fyne.CanvasObject {
	return sb.entry
}

// onSearch handles search query changes
func (sb *SearchBar) onSearch(query string) {
	sb.manager.SetSearch(query)
	if sb.list != nil {
		sb.list.UnselectAll()
	}
}