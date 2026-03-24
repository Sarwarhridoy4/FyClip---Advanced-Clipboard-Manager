// File: internal/app/app.go
package app

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/config"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/tray"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/ui"
)

//go:embed assets/icon.ico
var iconICO []byte

//go:embed assets/icon.png
var iconPNG []byte

// forcedVariantTheme is a custom theme that wraps the default theme
// but forces a specific variant (light or dark) instead of following system preference.
type forcedVariantTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

// Color overrides the default Color method to return colors for the forced variant.
func (f forcedVariantTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

// Variant returns the forced theme variant.
func (f forcedVariantTheme) Variant() fyne.ThemeVariant {
	return f.variant
}

// App represents the FyClip application
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	manager *clipboard.Manager
	mainUI  *ui.MainWindow
	tray    *tray.SystemTray
	configMgr *config.ConfigManager
}

// New creates a new FyClip application
func New() *App {
	a := &App{}

	// Initialize config manager first
	var err error
	a.configMgr, err = config.NewConfigManager()
	if err != nil {
		log.Printf("Failed to initialize config: %v, using defaults", err)
	}

	// Create Fyne app
	a.fyneApp = app.NewWithID("com.sarwar.fyclip")
	if a.fyneApp == nil {
		return nil
	}

	// Apply theme from config using custom theme to avoid deprecated LightTheme/DarkTheme
	if a.configMgr != nil {
		cfg := a.configMgr.Get()
		switch cfg.Theme {
		case "light":
			// Create a custom theme that forces light variant
			a.fyneApp.Settings().SetTheme(forcedVariantTheme{
				Theme:   theme.DefaultTheme(),
				variant: theme.VariantLight,
			})
		case "dark":
			// Create a custom theme that forces dark variant
			a.fyneApp.Settings().SetTheme(forcedVariantTheme{
				Theme:   theme.DefaultTheme(),
				variant: theme.VariantDark,
			})
		}
		// "system" theme uses default theme which follows system preference (no custom theme set)
	}

	icon := a.loadIcon()
	a.fyneApp.SetIcon(icon)

	// Create main window
	a.window = a.fyneApp.NewWindow("FyClip - Clipboard Manager")
	a.window.Resize(fyne.NewSize(900, 600))

	// CRITICAL: Linux dock icon fix
	a.window.SetIcon(icon)

	// Initialize clipboard manager
	a.manager, err = clipboard.NewManager(clipboard.Config{
		StoragePath: "",
		OnUpdate: func() {
			fyne.Do(func() {
				if a.mainUI != nil {
					a.mainUI.Refresh()
				}
				// Also refresh the tray menu to update recent items
				if a.tray != nil {
					a.tray.RefreshRecentMenu()
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
	a.tray = tray.New(a.fyneApp, a.window, a.manager, icon)
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
