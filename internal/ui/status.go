// File: internal/ui/status.go
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// StatusBar displays application status
type StatusBar struct {
	manager *clipboard.Manager
	label   *widget.Label
}

// NewStatusBar creates a new status bar
func NewStatusBar(manager *clipboard.Manager) *StatusBar {
	return &StatusBar{
		manager: manager,
		label:   widget.NewLabel(""),
	}
}

// Build returns the status widget
func (sb *StatusBar) Build() fyne.CanvasObject {
	sb.Refresh()
	return sb.label
}

// Refresh updates the status display
func (sb *StatusBar) Refresh() {
	total, pinned, filtered, lastCopied := sb.manager.GetStats()
	
	lastTime := "Never"
	if !lastCopied.IsZero() {
		lastTime = lastCopied.Format("15:04:05")
	}
	
	sb.label.SetText(fmt.Sprintf(
		"Total: %d | Pinned: %d | Showing: %d | Last copied: %s",
		total, pinned, filtered, lastTime,
	))
}