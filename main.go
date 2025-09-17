package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
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

// ---------------------- Clipboard Item Types ----------------------

type ClipboardItemType int

const (
	TypeText ClipboardItemType = iota
	TypeImage
)

type ClipboardItem struct {
	ID        string            `json:"id"`
	Type      ClipboardItemType `json:"type"`
	Content   string            `json:"content"`
	ImageData string            `json:"image_data,omitempty"` // Base64 encoded
	Timestamp time.Time         `json:"timestamp"`
	Pinned    bool              `json:"pinned"`
}

// ---------------------- Clipboard Manager ----------------------

const historyFile = "clipboard_history.json"

type ClipboardManager struct {
	mu               sync.RWMutex
	history          []ClipboardItem
	filtered         []ClipboardItem
	historyPath      string
	selectedIndex    int
	searchQuery      string
	historyList      *widget.List
	searchEntry      *widget.Entry
	statusLabel      *widget.Label
	previewEntry     *widget.Entry
	previewImage     *canvas.Image
	previewContainer *fyne.Container
	window           fyne.Window
	app              fyne.App
	shutdownChan     chan struct{}
	running          bool
	lastCopied       time.Time
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
		cm.history = []ClipboardItem{}
		return
	}
	if err := json.Unmarshal(data, &cm.history); err != nil {
		log.Printf("Warning: Could not parse history file: %v", err)
		cm.history = []ClipboardItem{}
	}
}

func (cm *ClipboardManager) saveHistory() {
	cm.mu.RLock()
	historyCopy := make([]ClipboardItem, len(cm.history))
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

	// Sort history: pinned items first, then by timestamp (newest first)
	sortedHistory := make([]ClipboardItem, len(cm.history))
	copy(sortedHistory, cm.history)

	// Separate pinned and unpinned items
	var pinnedItems, unpinnedItems []ClipboardItem
	for _, item := range sortedHistory {
		if item.Pinned {
			pinnedItems = append(pinnedItems, item)
		} else {
			unpinnedItems = append(unpinnedItems, item)
		}
	}

	// Combine: pinned first, then unpinned (newest first)
	sortedHistory = append(pinnedItems, unpinnedItems...)

	if cm.searchQuery == "" {
		cm.filtered = make([]ClipboardItem, len(sortedHistory))
		copy(cm.filtered, sortedHistory)
	} else {
		cm.filtered = cm.filtered[:0]
		query := strings.ToLower(cm.searchQuery)
		for _, item := range sortedHistory {
			if strings.Contains(strings.ToLower(item.Content), query) {
				cm.filtered = append(cm.filtered, item)
			}
		}
	}
	cm.selectedIndex = -1
}

func (cm *ClipboardManager) addToHistory(content string, itemType ClipboardItemType, imageData string) bool {
	if content == "" {
		return false
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check for duplicate
	var duplicate bool
	if itemType == TypeText {
		if len(cm.history) > 0 && cm.history[len(cm.history)-1].Content == content && cm.history[len(cm.history)-1].Type == itemType {
			duplicate = true
		}
	} else {
		if len(cm.history) > 0 && cm.history[len(cm.history)-1].ImageData == imageData && cm.history[len(cm.history)-1].Type == itemType {
			duplicate = true
		}
	}
	if duplicate {
		return false
	}

	// Remove existing duplicate
	for i := len(cm.history) - 1; i >= 0; i-- {
		if (itemType == TypeText && cm.history[i].Content == content) ||
			(itemType == TypeImage && cm.history[i].ImageData == imageData) &&
			cm.history[i].Type == itemType {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}

	// Create new item
	item := ClipboardItem{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      itemType,
		Content:   content,
		ImageData: imageData,
		Timestamp: time.Now(),
		Pinned:    false,
	}

	cm.history = append(cm.history, item)

	// Keep only last 1000 unpinned items (pinned items are kept)
	unpinnedCount := 0
	for i := len(cm.history) - 1; i >= 0; i-- {
		if !cm.history[i].Pinned {
			unpinnedCount++
			if unpinnedCount > 1000 {
				cm.history = append(cm.history[:i], cm.history[i+1:]...)
			}
		}
	}

	return true
}

func (cm *ClipboardManager) togglePin(index int) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if index < 0 || index >= len(cm.filtered) {
		return false
	}
	targetItem := cm.filtered[index]
	for i := range cm.history {
		if cm.history[i].ID == targetItem.ID {
			cm.history[i].Pinned = !cm.history[i].Pinned
			return true
		}
	}
	return false
}

func (cm *ClipboardManager) deleteSelected() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.selectedIndex < 0 || cm.selectedIndex >= len(cm.filtered) {
		return false
	}
	targetItem := cm.filtered[cm.selectedIndex]
	if targetItem.Pinned {
		fyne.Do(func() {
			cm.app.SendNotification(&fyne.Notification{
				Title:   "Cannot Delete",
				Content: "Pinned items cannot be deleted. Please unpin first.",
			})
		})
		return false
	}
	for i, item := range cm.history {
		if item.ID == targetItem.ID {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}
	cm.selectedIndex = -1
	// Clear preview
	fyne.Do(func() {
		if cm.previewEntry != nil {
			cm.previewEntry.SetText("")
			cm.previewEntry.SetPlaceHolder("Preview selected item...")
			cm.previewEntry.Show()
		}
		if cm.previewImage != nil {
			cm.previewImage.Resource = theme.BrokenImageIcon()
			cm.previewImage.Hide()
		}
	})
	return true
}

func (cm *ClipboardManager) clearHistory() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	// Keep only pinned items
	var pinnedItems []ClipboardItem
	for _, item := range cm.history {
		if item.Pinned {
			pinnedItems = append(pinnedItems, item)
		}
	}
	cm.history = pinnedItems
	cm.filtered = []ClipboardItem{}
	cm.selectedIndex = -1
	// Clear preview
	fyne.Do(func() {
		if cm.previewEntry != nil {
			cm.previewEntry.SetText("")
			cm.previewEntry.SetPlaceHolder("Preview selected item...")
			cm.previewEntry.Show()
		}
		if cm.previewImage != nil {
			cm.previewImage.Resource = theme.BrokenImageIcon()
			cm.previewImage.Hide()
		}
	})
}

func (cm *ClipboardManager) getFilteredCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.filtered)
}

func (cm *ClipboardManager) getFilteredItem(index int) ClipboardItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if index < 0 || index >= len(cm.filtered) {
		return ClipboardItem{}
	}
	return cm.filtered[index]
}

func (cm *ClipboardManager) getSelectedItem() ClipboardItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.selectedIndex < 0 || cm.selectedIndex >= len(cm.filtered) {
		return ClipboardItem{}
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

	// Use fyne.Do to ensure UI updates happen on the main thread
	fyne.Do(func() {
		if cm.historyList != nil && cm.running {
			cm.historyList.Refresh()
		}
		if cm.statusLabel != nil && cm.running {
			pinnedCount := 0
			cm.mu.RLock()
			for _, item := range cm.history {
				if item.Pinned {
					pinnedCount++
				}
			}
			historyLen := len(cm.history)
			filteredCount := len(cm.filtered)
			cm.mu.RUnlock()

			cm.statusLabel.SetText(fmt.Sprintf(
				"Total: %d | Pinned: %d | Showing: %d | Last copied: %s",
				historyLen,
				pinnedCount,
				filteredCount,
				cm.lastCopied.Format("15:04:05"),
			))
		}
		cm.updatePreview()
	})
}

func (cm *ClipboardManager) updatePreview() {
	if !cm.running || cm.selectedIndex < 0 {
		return
	}

	selectedItem := cm.getSelectedItem()
	if selectedItem.ID == "" {
		return
	}

	switch selectedItem.Type {
	case TypeText:
		if cm.previewEntry != nil {
			cm.previewEntry.SetText(selectedItem.Content)
			cm.previewEntry.Show()
		}
		if cm.previewImage != nil {
			cm.previewImage.Hide()
		}
	case TypeImage:
		if cm.previewEntry != nil {
			size := len(selectedItem.ImageData) * 3 / 4 // Approximate decoded size
			cm.previewEntry.SetText(fmt.Sprintf("Image copied at %s\nSize: ~%d bytes",
				selectedItem.Timestamp.Format("2006-01-02 15:04:05"),
				size))
			cm.previewEntry.Show()
		}
		if cm.previewImage != nil && selectedItem.ImageData != "" {
			imageBytes, err := base64.StdEncoding.DecodeString(selectedItem.ImageData)
			if err == nil {
				resource := fyne.NewStaticResource("clipboard_image", imageBytes)
				cm.previewImage.Resource = resource
				cm.previewImage.FillMode = canvas.ImageFillContain
				cm.previewImage.Show()
				cm.previewImage.Refresh()
			}
		}
	}
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
		var lastText string
		var lastImageHash string
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !cm.running {
					return
				}

				textBytes := clipboard.Read(clipboard.FmtText)
				imageBytes := clipboard.Read(clipboard.FmtImage)

				if len(textBytes) > 0 {
					content := string(textBytes)
					if content != lastText {
						lastText = content
						lastImageHash = ""
						if cm.addToHistory(content, TypeText, "") {
							cm.lastCopied = time.Now()
							cm.saveHistory()
							cm.updateFiltered()
							fyne.Do(func() {
								cm.refreshUI()
							})
						}
					}
				} else if len(imageBytes) > 0 {
					sum := sha256.Sum256(imageBytes)
					hash := hex.EncodeToString(sum[:])
					if hash != lastImageHash {
						lastImageHash = hash
						lastText = ""
						base64Data := base64.StdEncoding.EncodeToString(imageBytes)
						content := fmt.Sprintf("Image (%d bytes)", len(imageBytes))
						if cm.addToHistory(content, TypeImage, base64Data) {
							cm.lastCopied = time.Now()
							cm.saveHistory()
							cm.updateFiltered()
							fyne.Do(func() {
								cm.refreshUI()
							})
						}
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

// ---------------------- Custom List Item Widget ----------------------

func createListItem(item ClipboardItem, index int, cm *ClipboardManager) *fyne.Container {
	// Create pin button
	var pinIcon fyne.Resource
	if item.Pinned {
		pinIcon = theme.ConfirmIcon()
	} else {
		pinIcon = theme.RadioButtonIcon()
	}

	pinButton := widget.NewButtonWithIcon("", pinIcon, func() {
		go func() {
			if cm.togglePin(index) {
				cm.saveHistory()
				cm.updateFiltered()
				fyne.Do(func() {
					cm.refreshUI()
				})
			}
		}()
	})
	pinButton.Resize(fyne.NewSize(24, 24))

	// Create type icon
	var typeIcon fyne.Resource
	switch item.Type {
	case TypeText:
		typeIcon = theme.DocumentIcon()
	case TypeImage:
		typeIcon = theme.MediaPhotoIcon()
	}

	typeLabel := widget.NewIcon(typeIcon)

	// Create content label with padding
	displayText := item.Content
	if len(displayText) > 80 {
		displayText = displayText[:77] + "..."
	}
	displayText = strings.ReplaceAll(displayText, "\n", " ")

	contentLabel := widget.NewLabel(displayText)

	// Create timestamp label
	timeLabel := widget.NewLabel(item.Timestamp.Format("15:04"))
	timeLabel.TextStyle.Monospace = true

	// Create horizontal container with pin button first
	content := container.NewHBox(
		pinButton,
		typeLabel,
		contentLabel,
		widget.NewLabel(""), // Spacer
		timeLabel,
	)

	// Add padding around the content
	paddedContent := container.NewPadded(content)

	return container.NewBorder(nil, nil, nil, nil, paddedContent)
}

// ---------------------- Main ----------------------

func main() {
	myApp := app.NewWithID("com.sarwar.fyclip")
	myApp.Settings().SetTheme(theme.DarkTheme())
	myApp.SetIcon(loadIcon())

	err := clipboard.Init()
	if err != nil {
		log.Printf("Warning: Clipboard init failed: %v. Image support may be unavailable.", err)
	}

	myWindow := myApp.NewWindow("FYClip - Clipboard Manager")
	myWindow.Resize(fyne.NewSize(900, 600))

	cm := NewClipboardManager(myWindow, myApp)

	// Preview components
	previewText := widget.NewMultiLineEntry()
	previewText.SetPlaceHolder("Preview selected item...")
	cm.previewEntry = previewText

	previewImage := canvas.NewImageFromResource(theme.BrokenImageIcon())
	previewImage.Hide()
	cm.previewImage = previewImage

	// Preview container that can switch between text and image
	previewContainer := container.NewBorder(nil, nil, nil, nil,
		container.NewStack(previewText, previewImage))
	cm.previewContainer = previewContainer

	// History List with custom rendering
	historyList := widget.NewList(
		func() int { return cm.getFilteredCount() },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil, nil, widget.NewLabel("Loading..."))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			item := cm.getFilteredItem(i)
			if item.ID == "" {
				return
			}

			// Create new content for this item
			newContent := createListItem(item, i, cm)

			// Update the container with new item data
			if border := o.(*fyne.Container); border != nil {
				border.Objects = newContent.Objects
				border.Refresh()
			}
		},
	)
	cm.historyList = historyList

	historyList.OnSelected = func(id widget.ListItemID) {
		cm.setSelectedIndex(id)
		selectedItem := cm.getSelectedItem()
		if selectedItem.ID != "" {
			if selectedItem.Type == TypeText {
				clipboard.Write(clipboard.FmtText, []byte(selectedItem.Content))
			} else if selectedItem.Type == TypeImage && selectedItem.ImageData != "" {
				data, err := base64.StdEncoding.DecodeString(selectedItem.ImageData)
				if err == nil {
					clipboard.Write(clipboard.FmtImage, data)
				} else {
					log.Printf("Error decoding image for copy: %v", err)
				}
			}
			cm.lastCopied = time.Now()
			fyne.Do(func() {
				cm.refreshUI()
			})
		}
	}

	// Search
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search clipboard history...")
	searchEntry.OnChanged = func(query string) {
		cm.setSearchQuery(query)
		fyne.Do(func() {
			cm.refreshUI()
			historyList.UnselectAll()
		})
	}
	cm.searchEntry = searchEntry

	// Toolbar with pin toggle
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			selectedItem := cm.getSelectedItem()
			if selectedItem.ID != "" {
				if selectedItem.Type == TypeText {
					clipboard.Write(clipboard.FmtText, []byte(selectedItem.Content))
				} else if selectedItem.Type == TypeImage && selectedItem.ImageData != "" {
					data, err := base64.StdEncoding.DecodeString(selectedItem.ImageData)
					if err == nil {
						clipboard.Write(clipboard.FmtImage, data)
					} else {
						log.Printf("Error decoding image for copy: %v", err)
					}
				}
				cm.lastCopied = time.Now()
				fyne.Do(func() {
					cm.refreshUI()
				})
			}
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			if cm.selectedIndex >= 0 {
				go func() {
					if cm.togglePin(cm.selectedIndex) {
						cm.saveHistory()
						cm.updateFiltered()
						fyne.Do(func() {
							cm.refreshUI()
						})
					}
				}()
			}
		}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			go func() {
				if cm.deleteSelected() {
					cm.saveHistory()
					cm.updateFiltered()
					fyne.Do(func() {
						cm.refreshUI()
						historyList.UnselectAll()
					})
				}
			}()
		}),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			go func() {
				cm.clearHistory()
				cm.saveHistory()
				cm.updateFiltered()
				fyne.Do(func() {
					cm.refreshUI()
					historyList.UnselectAll()
				})
			}()
		}),
	)

	statusLabel := widget.NewLabel("")
	cm.statusLabel = statusLabel

	// Layout with adjusted proportions
	split := container.NewHSplit(historyList, previewContainer)
	split.SetOffset(0.5)

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
				cm.shutdown() // Stop clipboard monitoring
				myApp.Quit()  // Quit app
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