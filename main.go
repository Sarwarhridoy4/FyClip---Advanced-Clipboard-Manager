package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ---------------------- AutoStart Helpers ----------------------

func autoStartPath() string {
	usr, _ := user.Current()
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(usr.HomeDir, ".config", "autostart", "fyclip.desktop")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"),
			"Microsoft", "Windows", "Start Menu", "Programs", "Startup", "fyclip.lnk")
	case "darwin":
		return filepath.Join(usr.HomeDir, "Library", "LaunchAgents", "com.fyclip.plist")
	}
	return ""
}

func isAutoStartEnabled() bool {
	_, err := os.Stat(autoStartPath())
	return err == nil
}

func enableAutoStart(execPath string) error {
	path := autoStartPath()
	switch runtime.GOOS {
	case "linux":
		content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Name=FyClip
Comment=Clipboard Manager
`, execPath)
		os.MkdirAll(filepath.Dir(path), 0755)
		return os.WriteFile(path, []byte(content), 0644)
	case "windows":
		cmd := fmt.Sprintf(`$ws = New-Object -ComObject WScript.Shell;
$lnk = $ws.CreateShortcut("%s");
$lnk.TargetPath = "%s";
$lnk.Save()`, path, execPath)
		return exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", cmd).Run()
	case "darwin":
		content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.fyclip</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>`, execPath)
		os.MkdirAll(filepath.Dir(path), 0755)
		return os.WriteFile(path, []byte(content), 0644)
	}
	return nil
}

func disableAutoStart() error {
	return os.Remove(autoStartPath())
}

// ---------------------- Clipboard Manager ----------------------

const historyFile = "clipboard_history.json"

type ClipboardManager struct {
	mu            sync.RWMutex
	history       []string
	filtered      []string
	historyPath   string
	selectedIndex int
	searchQuery   string

	historyList  *widget.List
	searchEntry  *widget.Entry
	statusLabel  *widget.Label
	previewEntry *widget.Entry
	window       fyne.Window
	app          fyne.App

	shutdownChan chan struct{}
	running      bool
	lastCopied   time.Time
}

func NewClipboardManager(window fyne.Window, app fyne.App) *ClipboardManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v", err)
		homeDir = "."
	}
	cm := &ClipboardManager{
		historyPath:   filepath.Join(homeDir, historyFile),
		selectedIndex: -1,
		window:        window,
		app:           app,
		shutdownChan:  make(chan struct{}),
		running:       true,
	}
	cm.loadHistory()
	cm.updateFiltered()
	return cm
}

func (cm *ClipboardManager) loadHistory() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	data, err := os.ReadFile(cm.historyPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Could not read history file: %v", err)
		}
		cm.history = []string{}
		return
	}
	if err := json.Unmarshal(data, &cm.history); err != nil {
		log.Printf("Warning: Could not parse history file: %v", err)
		cm.history = []string{}
	}
}

func (cm *ClipboardManager) saveHistory() {
	cm.mu.RLock()
	historyCopy := make([]string, len(cm.history))
	copy(historyCopy, cm.history)
	cm.mu.RUnlock()
	go func() {
		data, err := json.MarshalIndent(historyCopy, "", "  ")
		if err != nil {
			log.Printf("Error marshaling history: %v", err)
			return
		}
		if err := os.WriteFile(cm.historyPath, data, 0644); err != nil {
			log.Printf("Error saving history: %v", err)
		}
	}()
}

func (cm *ClipboardManager) updateFiltered() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.searchQuery == "" {
		cm.filtered = make([]string, len(cm.history))
		copy(cm.filtered, cm.history)
	} else {
		cm.filtered = cm.filtered[:0]
		query := strings.ToLower(cm.searchQuery)
		for _, item := range cm.history {
			if strings.Contains(strings.ToLower(item), query) {
				cm.filtered = append(cm.filtered, item)
			}
		}
	}
	cm.selectedIndex = -1
}

func (cm *ClipboardManager) addToHistory(content string) bool {
	if content == "" || len(content) > 10000 {
		return false
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if len(cm.history) > 0 && cm.history[len(cm.history)-1] == content {
		return false
	}
	for i := len(cm.history) - 1; i >= 0; i-- {
		if cm.history[i] == content {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}
	cm.history = append(cm.history, content)
	if len(cm.history) > 1000 {
		cm.history = cm.history[1:]
	}
	return true
}

func (cm *ClipboardManager) deleteSelected() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.selectedIndex < 0 || cm.selectedIndex >= len(cm.filtered) {
		return false
	}
	targetItem := cm.filtered[cm.selectedIndex]
	for i, item := range cm.history {
		if item == targetItem {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}
	cm.selectedIndex = -1
	return true
}

func (cm *ClipboardManager) clearHistory() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.history = []string{}
	cm.filtered = []string{}
	cm.selectedIndex = -1
}

func (cm *ClipboardManager) getFilteredCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.filtered)
}

func (cm *ClipboardManager) getFilteredItem(index int) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if index < 0 || index >= len(cm.filtered) {
		return ""
	}
	item := cm.filtered[index]
	if len(item) > 100 {
		return item[:97] + "..."
	}
	return strings.ReplaceAll(item, "\n", " ")
}

func (cm *ClipboardManager) getSelectedItem() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.selectedIndex < 0 || cm.selectedIndex >= len(cm.filtered) {
		return ""
	}
	return cm.filtered[cm.selectedIndex]
}

func (cm *ClipboardManager) setSelectedIndex(index int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.selectedIndex = index
}

func (cm *ClipboardManager) setSearchQuery(query string) {
	cm.mu.Lock()
	cm.searchQuery = query
	cm.mu.Unlock()
	cm.updateFiltered()
}

func (cm *ClipboardManager) refreshUI() {
	if !cm.running {
		return
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		if cm.historyList != nil && cm.running {
			cm.historyList.Refresh()
		}
		if cm.statusLabel != nil && cm.running {
			cm.statusLabel.SetText(fmt.Sprintf(
				"Total: %d | Showing: %d | Last copied: %s",
				len(cm.history),
				cm.getFilteredCount(),
				cm.lastCopied.Format("15:04:05"),
			))
		}
		if cm.previewEntry != nil && cm.running && cm.selectedIndex >= 0 {
			cm.previewEntry.SetText(cm.getSelectedItem())
		}
	}()
}

func (cm *ClipboardManager) shutdown() {
	cm.mu.Lock()
	cm.running = false
	cm.mu.Unlock()
	select {
	case <-cm.shutdownChan:
	default:
		close(cm.shutdownChan)
	}
}

func (cm *ClipboardManager) startClipboardMonitor() {
	go func() {
		var lastContent string
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !cm.running {
					return
				}
				content := cm.window.Clipboard().Content()
				if content != "" && content != lastContent {
					lastContent = content
					if cm.addToHistory(content) {
						cm.lastCopied = time.Now()
						cm.saveHistory()
						cm.updateFiltered()
						cm.refreshUI()
					}
				}
			case <-cm.shutdownChan:
				return
			}
		}
	}()
}

func loadIcon() fyne.Resource {
	iconPath := "icon.png"
	if data, err := os.ReadFile(iconPath); err == nil {
		return fyne.NewStaticResource("icon.png", data)
	}
	log.Printf("Icon file 'icon.png' not found")
	return theme.ContentPasteIcon() // fallback
}

// ---------------------- Main ----------------------

func main() {
	myApp := app.NewWithID("com.example.fyclip")
	myApp.Settings().SetTheme(theme.DarkTheme())
	myApp.SetIcon(loadIcon())

	myWindow := myApp.NewWindow("FYClip - Clipboard Manager")
	myWindow.Resize(fyne.NewSize(800, 500))

	cm := NewClipboardManager(myWindow, myApp)

	// Preview
	preview := widget.NewMultiLineEntry()
	preview.SetPlaceHolder("Preview selected item...")
	cm.previewEntry = preview

	// History List
	historyList := widget.NewList(
		func() int { return cm.getFilteredCount() },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(cm.getFilteredItem(i))
		},
	)
	cm.historyList = historyList
	historyList.OnSelected = func(id widget.ListItemID) {
		cm.setSelectedIndex(id)
		if selectedItem := cm.getSelectedItem(); selectedItem != "" {
			myWindow.Clipboard().SetContent(selectedItem)
			cm.lastCopied = time.Now()
			cm.refreshUI()
		}
	}

	// Search
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search clipboard history...")
	searchEntry.OnChanged = func(query string) {
		cm.setSearchQuery(query)
		cm.refreshUI()
		historyList.UnselectAll()
	}
	cm.searchEntry = searchEntry

	// Toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			if selectedItem := cm.getSelectedItem(); selectedItem != "" {
				myWindow.Clipboard().SetContent(selectedItem)
				cm.lastCopied = time.Now()
				cm.refreshUI()
			}
		}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			if cm.deleteSelected() {
				cm.saveHistory()
				cm.updateFiltered()
				cm.refreshUI()
				historyList.UnselectAll()
			}
		}),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			cm.clearHistory()
			cm.saveHistory()
			cm.refreshUI()
			historyList.UnselectAll()
		}),
	)

	statusLabel := widget.NewLabel("")
	cm.statusLabel = statusLabel

	// Layout
	split := container.NewHSplit(historyList, preview)
	split.SetOffset(0.4)

	content := container.NewBorder(
		searchEntry,
		container.NewVBox(toolbar, statusLabel),
		nil, nil,
		split,
	)
	myWindow.SetContent(content)

	// Clipboard monitor
	cm.startClipboardMonitor()

	// System Tray Integration
if desk, ok := myApp.(desktop.App); ok {
    // AutoStart menu item
    autoStartItem := fyne.NewMenuItem("", nil)
    updateAutoStartLabel := func() {
        if isAutoStartEnabled() {
            autoStartItem.Label = "Disable AutoStart"
        } else {
            autoStartItem.Label = "Enable AutoStart"
        }
    }
    updateAutoStartLabel()

    autoStartItem.Action = func() {
        execPath, _ := os.Executable()
        if isAutoStartEnabled() {
            if err := disableAutoStart(); err != nil {
                log.Println("Error disabling autostart:", err)
            }
        } else {
            if err := enableAutoStart(execPath); err != nil {
                log.Println("Error enabling autostart:", err)
            }
        }
        updateAutoStartLabel()
        desk.SetSystemTrayMenu(buildTrayMenu(myWindow, myApp, autoStartItem))
    }

    // Build and set the tray menu
    trayMenu := fyne.NewMenu("FyClip",
        fyne.NewMenuItem("Show", func() {
            myWindow.Show()
        }),
        autoStartItem,
        fyne.NewMenuItem("Quit", func() {
            cm.shutdown()      // Stop clipboard monitoring
            myApp.Quit()       // Quit app
        }),
    )
    desk.SetSystemTrayMenu(trayMenu)

    // Use embedded icon for system tray
    desk.SetSystemTrayIcon(loadIcon())
}


	myWindow.SetCloseIntercept(func() { myWindow.Hide() })
	myWindow.ShowAndRun()
}

func buildTrayMenu(w fyne.Window, a fyne.App, autoStartItem *fyne.MenuItem) *fyne.Menu {
	return fyne.NewMenu("FyClip",
		fyne.NewMenuItem("Show", func() { w.Show() }),
		autoStartItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			a.Quit()
		}),
	)
}
