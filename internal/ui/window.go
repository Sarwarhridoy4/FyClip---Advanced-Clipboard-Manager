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
	mw.list = NewHistoryList(manager, mw.onItemSelected)
	mw.preview = NewPreviewPane(manager)
	mw.toolbar = NewToolbar(window, app, manager, mw.list)
	mw.search = NewSearchBar(manager, mw.list)
	mw.status = NewStatusBar(manager)

	// Create quick panel
	mw.quickPanel = NewQuickPanel(manager, window, mw.onQuickPaste)

	// Setup menu
	mw.setupMenu()

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

	return content
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

// setupMenu creates the application menu
func (mw *MainWindow) setupMenu() {
	aboutItem := fyne.NewMenuItem("About", func() {
		ShowAboutDialog(mw.window, mw.app)
	})
	
	// Add quick panel menu item
	quickPanelItem := fyne.NewMenuItem("Quick Paste", func() {
		if mw.quickPanel != nil {
			mw.quickPanel.Toggle()
		}
	})

	helpMenu := fyne.NewMenu("Help", aboutItem)
	viewMenu := fyne.NewMenu("View", quickPanelItem)
	mainMenu := fyne.NewMainMenu(viewMenu, helpMenu)
	mw.window.SetMainMenu(mainMenu)
}
