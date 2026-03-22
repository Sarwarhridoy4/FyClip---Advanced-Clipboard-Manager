// File: internal/ui/window.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// MainWindow represents the main application window
type MainWindow struct {
	window    fyne.Window
	app       fyne.App
	manager   *clipboard.Manager
	quickPanel *QuickPanel

	list    *HistoryList
	preview *PreviewPane
	toolbar *Toolbar
	search  *SearchBar
	status  *StatusBar
}

// NewMainWindow creates a new main window
func NewMainWindow(window fyne.Window, app fyne.App, manager *clipboard.Manager) *MainWindow {
	mw := &MainWindow{
		window:  window,
		app:     app,
		manager: manager,
	}

	// Create components
	mw.list = NewHistoryList(manager, mw.onItemSelected, app, window)
	mw.preview = NewPreviewPane(manager)
	mw.toolbar = NewToolbar(window, app, manager, mw.list)
	mw.search = NewSearchBar(manager, mw.list)
	mw.status = NewStatusBar(manager)

	// Create quick panel
	mw.quickPanel = NewQuickPanel(manager, window, mw.onQuickPaste)

	return mw
}

// Build constructs the UI layout
func (mw *MainWindow) Build() fyne.CanvasObject {
	// Split between list and preview
	split := container.NewHSplit(
		mw.list.Build(),
		mw.preview.Build(),
	)
	split.SetOffset(0.5)

	// Main layout
	content := container.NewBorder(
		mw.search.Build(),
		container.NewVBox(
			mw.toolbar.Build(),
			mw.status.Build(),
		),
		nil,
		nil,
		split,
	)

	// Setup keyboard shortcuts
	mw.setupShortcuts()

	return content
}

// setupShortcuts sets up the menu
func (mw *MainWindow) setupShortcuts() {
	// Edit menu items
	copyItem := fyne.NewMenuItem("Copy", func() {
		if mw.toolbar != nil {
			mw.toolbar.onCopy()
		}
	})

	deleteItem := fyne.NewMenuItem("Delete", func() {
		if mw.toolbar != nil {
			mw.toolbar.onDelete()
		}
	})

	searchItem := fyne.NewMenuItem("Search", func() {
		if mw.search != nil {
			mw.search.Focus(mw.window)
		}
	})

	backupItem := fyne.NewMenuItem("Backup", func() {
		if mw.toolbar != nil {
			mw.toolbar.onBackup()
		}
	})

	restoreItem := fyne.NewMenuItem("Restore", func() {
		if mw.toolbar != nil {
			mw.toolbar.onRestore()
		}
	})

	editMenu := fyne.NewMenu("Edit", copyItem, deleteItem, searchItem, backupItem, restoreItem)

	// View menu items
	quickPanelItem := fyne.NewMenuItem("Quick Paste", func() {
		if mw.quickPanel != nil {
			mw.quickPanel.Toggle()
		}
	})
	viewMenu := fyne.NewMenu("View", quickPanelItem)

	// Help menu items
	featuresItem := fyne.NewMenuItem("Features", func() {
		ShowFeaturesDialog(mw.window, mw.app)
	})

	aboutItem := fyne.NewMenuItem("About", func() {
		ShowAboutDialog(mw.window, mw.app)
	})

	helpMenu := fyne.NewMenu("Help", featuresItem, aboutItem)

	mainMenu := fyne.NewMainMenu(editMenu, viewMenu, helpMenu)
	mw.window.SetMainMenu(mainMenu)

	// Setup keyboard capture for navigation
	mw.setupKeyCapture()
}

// setupKeyCapture sets up key capture for vim-style navigation
func (mw *MainWindow) setupKeyCapture() {
	// Note: Keyboard navigation is handled through the list widget's built-in support.
	// This method can be extended later for custom keyboard shortcuts.
	// For now, the basic navigation (arrow keys, enter, delete) works out of the box.
}

// moveSelection moves the selection up or down
func (mw *MainWindow) moveSelection(direction int) {
	current := mw.manager.GetSelectedIndex()
	count := mw.manager.GetFilteredCount()
	if count == 0 {
		return
	}

	newIndex := current + direction
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= count {
		newIndex = count - 1
	}

	mw.manager.SetSelected(newIndex)
	if mw.list != nil {
		mw.list.Refresh()
	}
	if mw.preview != nil {
		mw.preview.Refresh()
	}
}

// moveToTop moves selection to the first item
func (mw *MainWindow) moveToTop() {
	mw.manager.SetSelected(0)
	if mw.list != nil {
		mw.list.Refresh()
	}
	if mw.preview != nil {
		mw.preview.Refresh()
	}
}

// moveToBottom moves selection to the last item
func (mw *MainWindow) moveToBottom() {
	count := mw.manager.GetFilteredCount()
	if count > 0 {
		mw.manager.SetSelected(count - 1)
		if mw.list != nil {
			mw.list.Refresh()
		}
		if mw.preview != nil {
			mw.preview.Refresh()
		}
	}
}

// Refresh updates all UI components
func (mw *MainWindow) Refresh() {
	if mw.list != nil {
		mw.list.Refresh()
	}
	if mw.preview != nil {
		mw.preview.Refresh()
	}
	if mw.toolbar != nil {
		mw.toolbar.Refresh()
	}
	if mw.status != nil {
		mw.status.Refresh()
	}
}

// onItemSelected handles item selection
func (mw *MainWindow) onItemSelected(index int) {
	mw.manager.SetSelected(index)
	if mw.preview != nil {
		mw.preview.Refresh()
	}
}

// onQuickPaste handles quick paste from the panel
func (mw *MainWindow) onQuickPaste(item clipboard.Item) {
	// Find the item in the manager and copy it
	items := mw.manager.GetFiltered()
	for i, it := range items {
		if it.ID == item.ID {
			mw.manager.CopyToClipboard(i)
			break
		}
	}
}

// ShowQuickPanel shows the quick paste panel
func (mw *MainWindow) ShowQuickPanel() {
	if mw.quickPanel != nil {
		mw.quickPanel.Show()
	}
}
