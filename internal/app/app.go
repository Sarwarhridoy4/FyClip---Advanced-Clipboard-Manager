// File: internal/app/app.go
package app

import (
	_ "embed"
	"fmt"
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/tray"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/ui"
)

//go:embed assets/icon.ico
var iconICO []byte

//go:embed assets/icon.png
var iconPNG []byte

// App represents the FyClip application
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	manager *clipboard.Manager
	mainUI  *ui.MainWindow
	tray    *tray.SystemTray
}

// New creates a new FyClip application
func New() *App {
	a := &App{}

	// Create Fyne app
	a.fyneApp = app.NewWithID("com.sarwar.fyclip")
	if a.fyneApp == nil {
		return nil
	}

	a.fyneApp.Settings().SetTheme(theme.DarkTheme())

	icon := a.loadIcon()
	a.fyneApp.SetIcon(icon)

	// Create main window
	a.window = a.fyneApp.NewWindow("FyClip - Clipboard Manager")
	a.window.Resize(fyne.NewSize(900, 600))

	// CRITICAL: Linux dock icon fix
	a.window.SetIcon(icon)

	// Initialize clipboard manager
	var err error
	a.manager, err = clipboard.NewManager(clipboard.Config{
		StoragePath: "",
		OnUpdate: func() {
			fyne.Do(func() {
				if a.mainUI != nil {
					a.mainUI.Refresh()
				}
			})
		},
		OnError: func(err error) {
			log.Printf("Manager error: %v", err)
			fyne.Do(func() {
				a.fyneApp.SendNotification(&fyne.Notification{
					Title:   "Error",
					Content: err.Error(),
				})
			})
		},
		OnInfo: func(message string) {
			fyne.Do(func() {
				a.fyneApp.SendNotification(&fyne.Notification{
					Title:   "FyClip",
					Content: message,
				})
			})
		},
	})
	if err != nil {
		log.Printf("Failed to create manager: %v", err)
		return nil
	}

	// Create main UI
	a.mainUI = ui.NewMainWindow(a.window, a.fyneApp, a.manager)
	a.window.SetContent(a.mainUI.Build())

	// Setup system tray
	a.tray = tray.New(a.fyneApp, a.window, a.manager)
	a.tray.Setup()

	// Hide window instead of exit
	a.window.SetCloseIntercept(func() {
		a.window.Hide()
	})

	return a
}

// Run starts the application
func (a *App) Run() error {
	if a == nil || a.window == nil {
		return fmt.Errorf("application not initialized")
	}

	defer func() {
		if a.manager != nil {
			a.manager.Shutdown()
		}
	}()

	a.window.ShowAndRun()
	return nil
}

// loadIcon loads the application icon (cross-platform safe)
func (a *App) loadIcon() fyne.Resource {
	switch runtime.GOOS {
	case "windows":
		if len(iconICO) > 0 {
			return fyne.NewStaticResource("icon.ico", iconICO)
		}
	default: // linux, darwin
		if len(iconPNG) > 0 {
			return fyne.NewStaticResource("icon.png", iconPNG)
		}
	}
	return theme.ContentPasteIcon()
}
