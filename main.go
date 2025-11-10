package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
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
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"

	_ "embed"
)

//go:embed icon.png
var iconBytes []byte

// ---------------------- AutoStart Helpers ----------------------

func autoStartPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Warning: Could not get current user: %v", err)
		return ""
	}
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(usr.HomeDir, ".config", "autostart", "fyclip.desktop")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			log.Printf("Warning: APPDATA environment variable not set")
			return ""
		}
		return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "fyclip.lnk")
	case "darwin":
		return filepath.Join(usr.HomeDir, "Library", "LaunchAgents", "com.fyclip.plist")
	}
	return ""
}

func isAutoStartEnabled() bool {
	path := autoStartPath()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func enableAutoStart(execPath string) error {
	path := autoStartPath()
	if path == "" {
		return fmt.Errorf("could not determine autostart path")
	}

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
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create autostart directory: %w", err)
		}
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
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create launch agents directory: %w", err)
		}
		return os.WriteFile(path, []byte(content), 0644)
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

func disableAutoStart() error {
	path := autoStartPath()
	if path == "" {
		return fmt.Errorf("could not determine autostart path")
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove autostart file: %w", err)
	}
	return nil
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
	ImageData string            `json:"image_data,omitempty"`
	ImageType string            `json:"image_type,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Pinned    bool              `json:"pinned"`
}

// ---------------------- Clipboard Manager ----------------------

const historyFile = "clipboard_history.json"

type ClipboardManager struct {
	mu                   sync.RWMutex
	history              []ClipboardItem
	filtered             []ClipboardItem
	historyPath          string
	selectedIndex        int
	searchQuery          string
	historyList          *widget.List
	searchEntry          *widget.Entry
	statusLabel          *widget.Label
	previewEntry         *widget.Entry
	previewImage         *canvas.Image
	previewContainer     *fyne.Container
	window               fyne.Window
	app                  fyne.App
	shutdownChan         chan struct{}
	running              bool
	lastCopied           time.Time
	useXclip             bool
	useWlclip            bool
	programmaticCopy     bool
	lastProgrammaticHash string
	clipboardAvailable   bool
}

func NewClipboardManager(window fyne.Window, app fyne.App) *ClipboardManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v", err)
		homeDir = "."
	}
	cm := &ClipboardManager{
		historyPath:        filepath.Join(homeDir, historyFile),
		selectedIndex:      -1,
		window:             window,
		app:                app,
		shutdownChan:       make(chan struct{}),
		running:            true,
		clipboardAvailable: true,
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

	sortedHistory := make([]ClipboardItem, len(cm.history))
	copy(sortedHistory, cm.history)

	var pinnedItems, unpinnedItems []ClipboardItem
	for _, item := range sortedHistory {
		if item.Pinned {
			pinnedItems = append(pinnedItems, item)
		} else {
			unpinnedItems = append(unpinnedItems, item)
		}
	}

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

func (cm *ClipboardManager) addToHistory(content string, itemType ClipboardItemType, imageData string, imageType string) bool {
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
		if (itemType == TypeText && cm.history[i].Content == content && cm.history[i].Type == itemType) ||
			(itemType == TypeImage && cm.history[i].ImageData == imageData && cm.history[i].Type == itemType) {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}

	item := ClipboardItem{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      itemType,
		Content:   content,
		ImageData: imageData,
		ImageType: imageType,
		Timestamp: time.Now(),
		Pinned:    false,
	}

	cm.history = append(cm.history, item)

	// Keep only last 1000 unpinned items
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
			if cm.app != nil {
				cm.app.SendNotification(&fyne.Notification{
					Title:   "Cannot Delete",
					Content: "Pinned items cannot be deleted. Please unpin first.",
				})
			}
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
	var pinnedItems []ClipboardItem
	for _, item := range cm.history {
		if item.Pinned {
			pinnedItems = append(pinnedItems, item)
		}
	}
	cm.history = pinnedItems
	cm.filtered = []ClipboardItem{}
	cm.selectedIndex = -1
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

func (cm *ClipboardManager) saveImageAs(filename string, format string) {
	if cm.selectedIndex < 0 {
		return
	}
	selectedItem := cm.getSelectedItem()
	if selectedItem.Type != TypeImage || selectedItem.ImageData == "" {
		fyne.Do(func() {
			if cm.app != nil {
				cm.app.SendNotification(&fyne.Notification{
					Title:   "Error",
					Content: "No image selected or image data is empty.",
				})
			}
		})
		return
	}

	data, err := base64.StdEncoding.DecodeString(selectedItem.ImageData)
	if err != nil {
		log.Printf("Error decoding image data: %v", err)
		fyne.Do(func() {
			if cm.app != nil {
				cm.app.SendNotification(&fyne.Notification{
					Title:   "Error",
					Content: "Failed to decode image data.",
				})
			}
		})
		return
	}

	img, _, err := image.Decode(strings.NewReader(string(data)))
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		fyne.Do(func() {
			if cm.app != nil {
				cm.app.SendNotification(&fyne.Notification{
					Title:   "Error",
					Content: "Failed to decode image format.",
				})
			}
		})
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		fyne.Do(func() {
			if cm.app != nil {
				cm.app.SendNotification(&fyne.Notification{
					Title:   "Error",
					Content: "Failed to create file.",
				})
			}
		})
		return
	}
	defer file.Close()

	switch format {
	case "png":
		if err := png.Encode(file, img); err != nil {
			log.Printf("Error encoding PNG: %v", err)
			fyne.Do(func() {
				if cm.app != nil {
					cm.app.SendNotification(&fyne.Notification{
						Title:   "Error",
						Content: "Failed to save image as PNG.",
					})
				}
			})
			return
		}
	case "jpeg", "jpg":
		if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 95}); err != nil {
			log.Printf("Error encoding JPEG: %v", err)
			fyne.Do(func() {
				if cm.app != nil {
					cm.app.SendNotification(&fyne.Notification{
						Title:   "Error",
						Content: "Failed to save image as JPEG.",
					})
				}
			})
			return
		}
	}

	fyne.Do(func() {
		if cm.app != nil {
			cm.app.SendNotification(&fyne.Notification{
				Title:   "Success",
				Content: fmt.Sprintf("Image saved as %s", filepath.Base(filename)),
			})
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
			size := len(selectedItem.ImageData) * 3 / 4
			imageTypeStr := strings.ToUpper(selectedItem.ImageType)
			if imageTypeStr == "" {
				imageTypeStr = "PNG"
			}
			cm.previewEntry.SetText(fmt.Sprintf("%s Image copied at %s\nSize: ~%d bytes",
				imageTypeStr,
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

func (cm *ClipboardManager) readClipboardText() []byte {
	if !cm.clipboardAvailable {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in readClipboardText: %v", r)
		}
	}()

	if cm.useWlclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "wl-paste", "-n")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		return out
	} else if cm.useXclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		return out
	}
	
	return clipboard.Read(clipboard.FmtText)
}

func (cm *ClipboardManager) readClipboardImage() ([]byte, string) {
	if !cm.clipboardAvailable {
		return nil, ""
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in readClipboardImage: %v", r)
		}
	}()

	if cm.useWlclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "wl-paste", "-t", "image/png", "-n")
		out, err := cmd.Output()
		if err == nil && len(out) > 0 {
			return out, "png"
		}
		
		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		cmd2 := exec.CommandContext(ctx2, "wl-paste", "-t", "image/jpeg", "-n")
		out, err = cmd2.Output()
		if err == nil && len(out) > 0 {
			return out, "jpeg"
		}
		return nil, ""
	} else if cm.useXclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", "image/png", "-o")
		out, err := cmd.Output()
		if err == nil && len(out) > 0 {
			return out, "png"
		}
		
		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		cmd2 := exec.CommandContext(ctx2, "xclip", "-selection", "clipboard", "-t", "image/jpeg", "-o")
		out, err = cmd2.Output()
		if err == nil && len(out) > 0 {
			return out, "jpeg"
		}
		return nil, ""
	}

	imgData := clipboard.Read(clipboard.FmtImage)
	if len(imgData) > 0 {
		imgType := "png"
		if len(imgData) > 8 {
			if imgData[0] == 0x89 && imgData[1] == 0x50 && imgData[2] == 0x4E && imgData[3] == 0x47 {
				imgType = "png"
			} else if imgData[0] == 0xFF && imgData[1] == 0xD8 {
				imgType = "jpeg"
			}
		}
		return imgData, imgType
	}
	return nil, ""
}

func (cm *ClipboardManager) writeClipboardText(data []byte) {
	if !cm.clipboardAvailable {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in writeClipboardText: %v", r)
		}
	}()

	cm.mu.Lock()
	cm.programmaticCopy = true
	sum := sha256.Sum256(data)
	cm.lastProgrammaticHash = hex.EncodeToString(sum[:])
	cm.mu.Unlock()

	if cm.useWlclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "wl-copy")
		cmd.Stdin = strings.NewReader(string(data))
		_ = cmd.Run()
	} else if cm.useXclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "xclip", "-i", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(string(data))
		_ = cmd.Run()
	} else {
		clipboard.Write(clipboard.FmtText, data)
	}

	time.AfterFunc(200*time.Millisecond, func() {
		cm.mu.Lock()
		cm.programmaticCopy = false
		cm.lastProgrammaticHash = ""
		cm.mu.Unlock()
	})
}

func (cm *ClipboardManager) writeClipboardImage(data []byte) {
	if !cm.clipboardAvailable {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in writeClipboardImage: %v", r)
		}
	}()

	cm.mu.Lock()
	cm.programmaticCopy = true
	sum := sha256.Sum256(data)
	cm.lastProgrammaticHash = hex.EncodeToString(sum[:])
	cm.mu.Unlock()

	if cm.useWlclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "wl-copy", "--type", "image/png")
		cmd.Stdin = strings.NewReader(string(data))
		_ = cmd.Run()
	} else if cm.useXclip {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", "image/png", "-i")
		cmd.Stdin = strings.NewReader(string(data))
		_ = cmd.Run()
	} else {
		clipboard.Write(clipboard.FmtImage, data)
	}

	time.AfterFunc(200*time.Millisecond, func() {
		cm.mu.Lock()
		cm.programmaticCopy = false
		cm.lastProgrammaticHash = ""
		cm.mu.Unlock()
	})
}

func (cm *ClipboardManager) startClipboardMonitor() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in clipboard monitor: %v", r)
				// Restart monitor after panic
				time.Sleep(1 * time.Second)
				if cm.running {
					cm.startClipboardMonitor()
				}
			}
		}()

		var lastText string
		var lastImageHash string
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !cm.running {
					return
				}

				cm.mu.RLock()
				isProgrammatic := cm.programmaticCopy
				programmaticHash := cm.lastProgrammaticHash
				cm.mu.RUnlock()

				if isProgrammatic {
					continue
				}

				textBytes := cm.readClipboardText()
				var imageBytes []byte
				var imageType string

				if len(textBytes) == 0 {
					imageBytes, imageType = cm.readClipboardImage()
				}

				if len(textBytes) > 0 {
					content := string(textBytes)
					sum := sha256.Sum256(textBytes)
					hash := hex.EncodeToString(sum[:])

					if content != lastText && hash != programmaticHash {
						lastText = content
						lastImageHash = ""
						if cm.addToHistory(content, TypeText, "", "") {
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

					if hash != lastImageHash && hash != programmaticHash {
						lastImageHash = hash
						lastText = ""
						base64Data := base64.StdEncoding.EncodeToString(imageBytes)
						content := fmt.Sprintf("Image (%d bytes)", len(imageBytes))
						if cm.addToHistory(content, TypeImage, base64Data, imageType) {
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
	if len(iconBytes) > 0 {
		return fyne.NewStaticResource("icon.png", iconBytes)
	}
	log.Printf("Embedded icon not found, falling back to theme icon")
	return theme.ContentPasteIcon()
}

// ---------------------- Custom List Item Widget ----------------------

func createListItem(item ClipboardItem, index int, cm *ClipboardManager) *fyne.Container {
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

	var typeIcon fyne.Resource
	switch item.Type {
	case TypeText:
		typeIcon = theme.DocumentIcon()
	case TypeImage:
		typeIcon = theme.MediaPhotoIcon()
	}

	typeLabel := widget.NewIcon(typeIcon)

	displayText := item.Content
	if len(displayText) > 80 {
		displayText = displayText[:77] + "..."
	}
	displayText = strings.ReplaceAll(displayText, "\n", " ")

	contentLabel := widget.NewLabel(displayText)

	timeLabel := widget.NewLabel(item.Timestamp.Format("15:04"))
	timeLabel.TextStyle.Monospace = true

	content := container.NewHBox(
		pinButton,
		typeLabel,
		contentLabel,
		widget.NewLabel(""),
		timeLabel,
	)

	paddedContent := container.NewPadded(content)

	return container.NewBorder(nil, nil, nil, nil, paddedContent)
}

// ---------------------- About Window ----------------------

func createAboutWindow(app fyne.App) fyne.Window {
	aboutWindow := app.NewWindow("About FyClip")
	aboutWindow.Resize(fyne.NewSize(400, 300))

	title := widget.NewLabel("FyClip - Advanced Clipboard Manager")
	title.TextStyle.Bold = true

	hyperlink := widget.NewHyperlink("github.com/Sarwarhridoy4", nil)
	if err := hyperlink.SetURLFromString("https://github.com/Sarwarhridoy4"); err != nil {
		log.Printf("Warning: Failed to set hyperlink URL: %v", err)
	}

	content := container.NewVBox(
		title,
		widget.NewLabel(""),
		widget.NewLabel("Developed by: Sarwar Hossain"),
		widget.NewLabel("Email: sarwarhridoy4@gmail.com"),
		hyperlink,
	)

	aboutWindow.SetContent(content)
	return aboutWindow
}

// ---------------------- Helper for Showing Popup with Auto-Dismiss ----------------------

func showActionPopup(window fyne.Window, message string) {
	if window == nil || window.Canvas() == nil {
		log.Printf("Warning: Cannot show popup - window or canvas is nil")
		return
	}

	popupContent := widget.NewCard("", "", widget.NewLabel(message))
	popup := widget.NewPopUp(popupContent, window.Canvas())

	contentPos := window.Content().Position()
	popupPos := contentPos.Add(fyne.NewPos(10, 40))
	popup.Move(popupPos)
	popup.Resize(fyne.NewSize(200, 40))
	popup.Show()

	time.AfterFunc(2*time.Second, func() {
		fyne.Do(func() {
			popup.Hide()
		})
	})
}

// ---------------------- Copy Item to Clipboard ----------------------

func (cm *ClipboardManager) copyItemToClipboard(item ClipboardItem) {
	if item.ID == "" {
		return
	}

	if item.Type == TypeText {
		cm.writeClipboardText([]byte(item.Content))
	} else if item.Type == TypeImage && item.ImageData != "" {
		data, err := base64.StdEncoding.DecodeString(item.ImageData)
		if err == nil {
			cm.writeClipboardImage(data)
		} else {
			log.Printf("Error decoding image for copy: %v", err)
			fyne.Do(func() {
				if cm.app != nil {
					cm.app.SendNotification(&fyne.Notification{
						Title:   "Error",
						Content: "Failed to copy image to clipboard.",
					})
				}
			})
			return
		}
	}
	cm.lastCopied = time.Now()
}

// ---------------------- Build Tray Menu ----------------------

func buildTrayMenu(w fyne.Window, a fyne.App, autoStartItem *fyne.MenuItem, cm *ClipboardManager) *fyne.Menu {
	return fyne.NewMenu("FyClip",
		fyne.NewMenuItem("Show", func() {
			if w != nil {
				w.Show()
			}
		}),
		autoStartItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			if cm != nil {
				cm.shutdown()
			}
			if a != nil {
				a.Quit()
			}
		}),
	)
}

// ---------------------- Main ----------------------

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Fatal panic in main: %v", r)
			time.Sleep(3 * time.Second) // Give time to see the error
		}
	}()

	myApp := app.NewWithID("com.sarwar.fyclip")
	if myApp == nil {
		log.Fatal("Failed to create application")
		return
	}

	myApp.Settings().SetTheme(theme.DarkTheme())
	myApp.SetIcon(loadIcon())

	// Initialize clipboard with error handling
	err := clipboard.Init()
	useFallback := runtime.GOOS == "linux" && err != nil
	if err != nil {
		log.Printf("Warning: Native clipboard init failed: %v. Attempting fallback methods.", err)
	}

	myWindow := myApp.NewWindow("FyClip - Clipboard Manager")
	if myWindow == nil {
		log.Fatal("Failed to create window")
		return
	}
	myWindow.Resize(fyne.NewSize(900, 600))

	cm := NewClipboardManager(myWindow, myApp)
	if cm == nil {
		log.Fatal("Failed to create clipboard manager")
		return
	}

	// Setup clipboard fallback methods for Linux
	if useFallback {
		sessionType := os.Getenv("XDG_SESSION_TYPE")
		if sessionType == "wayland" || sessionType == "" {
			_, wlErr := exec.LookPath("wl-paste")
			if wlErr == nil {
				cm.useWlclip = true
				log.Println("Using wl-clipboard fallback for Wayland.")
			} else {
				log.Println("Warning: wl-clipboard not found. Install wl-clipboard package for clipboard support on Wayland.")
				cm.clipboardAvailable = false
			}
		} else {
			_, xErr := exec.LookPath("xclip")
			if xErr == nil {
				cm.useXclip = true
				log.Println("Using xclip fallback for X11.")
			} else {
				log.Println("Warning: xclip not found. Install xclip package for clipboard support on X11.")
				cm.clipboardAvailable = false
			}
		}
	}

	// Preview components
	previewText := widget.NewMultiLineEntry()
	previewText.SetPlaceHolder("Preview selected item...")
	previewText.Wrapping = fyne.TextWrapWord
	cm.previewEntry = previewText

	previewImage := canvas.NewImageFromResource(theme.BrokenImageIcon())
	previewImage.Hide()
	cm.previewImage = previewImage

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

			newContent := createListItem(item, i, cm)

			if border, ok := o.(*fyne.Container); ok && border != nil {
				border.Objects = newContent.Objects
				border.Refresh()
			}
		},
	)
	cm.historyList = historyList

	// Update preview on selection
	historyList.OnSelected = func(id widget.ListItemID) {
		cm.setSelectedIndex(id)
		fyne.Do(func() {
			cm.updatePreview()
		})
	}

	// Search
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search clipboard history...")
	searchEntry.OnChanged = func(query string) {
		cm.setSearchQuery(query)
		fyne.Do(func() {
			cm.refreshUI()
			if historyList != nil {
				historyList.UnselectAll()
			}
		})
	}
	cm.searchEntry = searchEntry

	// Toolbar buttons
	copyButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		selectedItem := cm.getSelectedItem()
		if selectedItem.ID != "" {
			cm.copyItemToClipboard(selectedItem)
			fyne.Do(func() {
				cm.refreshUI()
				showActionPopup(myWindow, "Copied to clipboard!")
			})
		} else {
			showActionPopup(myWindow, "No item selected!")
		}
	})

	pinButton := widget.NewButtonWithIcon("Pin/Unpin", theme.ViewRefreshIcon(), func() {
		if cm.selectedIndex >= 0 {
			go func() {
				if cm.togglePin(cm.selectedIndex) {
					cm.saveHistory()
					cm.updateFiltered()
					item := cm.getSelectedItem()
					pinStatus := "Pinned"
					if !item.Pinned {
						pinStatus = "Unpinned"
					}
					fyne.Do(func() {
						cm.refreshUI()
						showActionPopup(myWindow, pinStatus+" item!")
					})
				}
			}()
		} else {
			showActionPopup(myWindow, "No item selected!")
		}
	})

	deleteButton := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if cm.selectedIndex < 0 {
			showActionPopup(myWindow, "No item selected!")
			return
		}

		go func() {
			if cm.deleteSelected() {
				cm.saveHistory()
				cm.updateFiltered()
				fyne.Do(func() {
					cm.refreshUI()
					if historyList != nil {
						historyList.UnselectAll()
					}
					showActionPopup(myWindow, "Deleted selected item!")
				})
			}
		}()
	})

	clearButton := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		dialog.ShowConfirm("Clear History",
			"Are you sure you want to clear all unpinned items?",
			func(confirmed bool) {
				if confirmed {
					go func() {
						cm.clearHistory()
						cm.saveHistory()
						cm.updateFiltered()
						fyne.Do(func() {
							cm.refreshUI()
							if historyList != nil {
								historyList.UnselectAll()
							}
							showActionPopup(myWindow, "Cleared clipboard history!")
						})
					}()
				}
			}, myWindow)
	})

	saveButton := widget.NewButtonWithIcon("Save Image", theme.DocumentSaveIcon(), func() {
		if cm.selectedIndex < 0 || cm.getSelectedItem().Type != TypeImage {
			showActionPopup(myWindow, "Please select an image!")
			return
		}
		fyne.Do(func() {
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil || writer == nil {
					return
				}
				defer writer.Close()
				uri := writer.URI()
				filename := uri.Path()

				// Determine format from filename or default to PNG
				format := "png"
				lowerFilename := strings.ToLower(filename)
				if strings.HasSuffix(lowerFilename, ".jpg") || strings.HasSuffix(lowerFilename, ".jpeg") {
					format = "jpeg"
				} else if !strings.HasSuffix(lowerFilename, ".png") {
					filename += ".png"
				}

				cm.saveImageAs(filename, format)
				showActionPopup(myWindow, "Image saved successfully!")
			}, myWindow)
		})
	})

	buttonsBox := container.NewHBox(
		copyButton,
		pinButton,
		deleteButton,
		clearButton,
		saveButton,
	)

	// Menu bar
	aboutItem := fyne.NewMenuItem("About", func() {
		aboutWindow := createAboutWindow(myApp)
		aboutWindow.Show()
	})

	helpMenu := fyne.NewMenu("Help", aboutItem)
	menu := fyne.NewMainMenu(helpMenu)
	myWindow.SetMainMenu(menu)

	statusLabel := widget.NewLabel("")
	cm.statusLabel = statusLabel

	// Layout
	split := container.NewHSplit(historyList, previewContainer)
	split.SetOffset(0.5)

	content := container.NewBorder(
		searchEntry,
		container.NewVBox(buttonsBox, statusLabel),
		nil, nil,
		split,
	)

	myWindow.SetContent(content)

	// Initial UI refresh
	cm.refreshUI()

	// Start clipboard monitor
	cm.startClipboardMonitor()

	// System tray setup
	if desk, ok := myApp.(desktop.App); ok {
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
			execPath, err := os.Executable()
			if err != nil {
				log.Printf("Error getting executable path: %v", err)
				return
			}
			if isAutoStartEnabled() {
				if err := disableAutoStart(); err != nil {
					log.Printf("Error disabling autostart: %v", err)
				}
			} else {
				if err := enableAutoStart(execPath); err != nil {
					log.Printf("Error enabling autostart: %v", err)
				}
			}
			updateAutoStartLabel()
			desk.SetSystemTrayMenu(buildTrayMenu(myWindow, myApp, autoStartItem, cm))
		}

		trayMenu := buildTrayMenu(myWindow, myApp, autoStartItem, cm)
		desk.SetSystemTrayMenu(trayMenu)
		desk.SetSystemTrayIcon(loadIcon())
	}

	myWindow.SetCloseIntercept(func() {
		myWindow.Hide()
	})

	myWindow.ShowAndRun()

	// Cleanup on exit
	cm.shutdown()
}