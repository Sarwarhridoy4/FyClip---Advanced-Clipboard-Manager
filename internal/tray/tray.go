// File: internal/tray/tray.go
package tray

import (
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/platform"
)

// SystemTray manages the system tray integration
type SystemTray struct {
	app           fyne.App
	window        fyne.Window
	manager       *clipboard.Manager
	autoStartMgr  *platform.AutoStart
	icon          fyne.Resource

	menu            *fyne.Menu
	autoStartItem   *fyne.MenuItem
	pauseItem       *fyne.MenuItem
	recentMenu     *fyne.Menu
	restartCount   int
}

// New creates a new system tray
func New(app fyne.App, window fyne.Window, manager *clipboard.Manager, icon fyne.Resource) *SystemTray {
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Failed to get executable path: %v", err)
		execPath = ""
	}

	return &SystemTray{
		app:          app,
		window:       window,
		manager:      manager,
		autoStartMgr: platform.NewAutoStart(execPath),
		icon:         icon,
	}
}

// Setup configures the system tray
func (st *SystemTray) Setup() {
	desk, ok := st.app.(desktop.App)
	if !ok {
		log.Println("System tray not supported on this platform")
		return
	}

	// Create autostart menu item
	st.autoStartItem = fyne.NewMenuItem("", st.onAutoStartToggle)
	st.updateAutoStartLabel()
	st.pauseItem = fyne.NewMenuItem("", st.onPauseToggle)
	st.updatePauseLabel()

	// Create recent items submenu
	st.recentMenu = fyne.NewMenu("Recent", st.buildRecentMenuItems()...)

	// Build menu
	st.menu = st.buildMenu()

	// Set system tray
	desk.SetSystemTrayMenu(st.menu)
	if icon := st.loadIcon(); icon != nil {
		desk.SetSystemTrayIcon(icon)
	}
}

// buildMenu creates the tray menu
func (st *SystemTray) buildMenu() *fyne.Menu {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Show", st.onShow),
		fyne.NewMenuItemSeparator(),
		st.pauseItem,
		fyne.NewMenuItem("Clear History", st.onClearHistory),
		fyne.NewMenuItemSeparator(),
	}

	// Add recent items as regular menu items (Fyne v2.4 compatible)
	if st.recentMenu != nil && len(st.recentMenu.Items) > 0 {
		items = append(items, fyne.NewMenuItemSeparator())
		items = append(items, st.buildRecentMenuItems()...)
	}

	items = append(items,
		fyne.NewMenuItemSeparator(),
		st.autoStartItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", st.onQuit),
	)

	return fyne.NewMenu("FyClip", items...)
}

// buildRecentMenuItems creates recent items menu items
func (st *SystemTray) buildRecentMenuItems() []*fyne.MenuItem {
	items := []*fyne.MenuItem{}

	if st.manager == nil {
		return append(items, fyne.NewMenuItem("No items", nil))
	}

	history := st.manager.GetFiltered()
	if len(history) == 0 {
		return append(items, fyne.NewMenuItem("No items", nil))
	}

	maxItems := 5
	if len(history) < maxItems {
		maxItems = len(history)
	}

	for i := 0; i < maxItems; i++ {
		item := history[i]
		text := item.DisplayText(40)
		if item.Pinned {
			text = "★ " + text
		}

		index := i
		menuItem := fyne.NewMenuItem(text, func() {
			if st.manager != nil {
				_ = st.manager.CopyToClipboard(index)
			}
		})
		items = append(items, menuItem)
	}

	if len(items) == 0 {
		items = append(items, fyne.NewMenuItem("No items", nil))
	}

	return items
}

// RefreshRecentMenu updates the recent items submenu
func (st *SystemTray) RefreshRecentMenu() {
	if st.recentMenu == nil {
		return
	}
	st.recentMenu.Items = st.buildRecentMenuItems()
	st.refreshMenu()
}

// onClearHistory handles clear history menu item
func (st *SystemTray) onClearHistory() {
	if st.manager != nil {
		st.manager.ClearUnpinned()
		st.manager.SaveHistory()
	}
}

// onShow handles show menu item
func (st *SystemTray) onShow() {
	if st.window != nil {
		st.window.Show()
		st.window.RequestFocus()
	}
}

// onAutoStartToggle handles autostart toggle
func (st *SystemTray) onAutoStartToggle() {
	if st.autoStartMgr.IsEnabled() {
		if err := st.autoStartMgr.Disable(); err != nil {
			log.Printf("Failed to disable autostart: %v", err)
			return
		}
	} else {
		if err := st.autoStartMgr.Enable(); err != nil {
			log.Printf("Failed to enable autostart: %v", err)
			return
		}
	}

	st.updateAutoStartLabel()
	st.refreshMenu()
}

func (st *SystemTray) onPauseToggle() {
	if st.manager == nil {
		return
	}
	if st.manager.IsMonitoringPaused() {
		st.manager.ResumeMonitoring()
	} else {
		st.manager.PauseMonitoringFor(5 * time.Minute)
	}
	st.updatePauseLabel()
	st.refreshMenu()
}

// onQuit handles quit menu item
func (st *SystemTray) onQuit() {
	if st.manager != nil {
		st.manager.Shutdown()
	}
	if st.app != nil {
		st.app.Quit()
	}
}

// updateAutoStartLabel updates the autostart menu item label
func (st *SystemTray) updateAutoStartLabel() {
	if st.autoStartItem == nil {
		return
	}

	if st.autoStartMgr.IsEnabled() {
		st.autoStartItem.Label = "✓ Auto-Start"
	} else {
		st.autoStartItem.Label = "Auto-Start"
	}
}

func (st *SystemTray) updatePauseLabel() {
	if st.pauseItem == nil {
		return
	}
	if st.manager != nil && st.manager.IsMonitoringPaused() {
		st.pauseItem.Label = "▶ Resume Monitoring"
	} else {
		st.pauseItem.Label = "⏸ Pause Monitoring"
	}
}

// refreshMenu rebuilds and updates the tray menu
func (st *SystemTray) refreshMenu() {
	desk, ok := st.app.(desktop.App)
	if !ok {
		return
	}

	st.menu = st.buildMenu()
	desk.SetSystemTrayMenu(st.menu)
}

// loadIcon loads the application icon
func (st *SystemTray) loadIcon() fyne.Resource {
	// Try to load from embedded icon
	// This should match the icon in app.go
	// Since we can't directly access the embedded icons from app package here,
	// we'll use the icon that was passed in during initialization
	if st.icon != nil {
		return st.icon
	}
	// Fallback to default icon
	return nil
}
