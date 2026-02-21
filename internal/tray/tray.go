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
	app          fyne.App
	window       fyne.Window
	manager      *clipboard.Manager
	autoStartMgr *platform.AutoStart

	menu          *fyne.Menu
	autoStartItem *fyne.MenuItem
	pauseItem     *fyne.MenuItem
}

// New creates a new system tray
func New(app fyne.App, window fyne.Window, manager *clipboard.Manager) *SystemTray {
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
	return fyne.NewMenu("FyClip",
		fyne.NewMenuItem("Show", st.onShow),
		st.pauseItem,
		st.autoStartItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", st.onQuit),
	)
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
		st.autoStartItem.Label = "Disable AutoStart"
	} else {
		st.autoStartItem.Label = "Enable AutoStart"
	}
}

func (st *SystemTray) updatePauseLabel() {
	if st.pauseItem == nil {
		return
	}
	if st.manager != nil && st.manager.IsMonitoringPaused() {
		st.pauseItem.Label = "Resume Monitoring"
	} else {
		st.pauseItem.Label = "Pause Monitoring (5m)"
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
	return nil // Will use default if nil
}
