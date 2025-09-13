package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// File to store history
const historyFile = "clipboard_history.json"

// ClipboardManager manages the clipboard history with thread safety
type ClipboardManager struct {
	mu            sync.RWMutex
	history       []string
	filtered      []string
	historyPath   string
	selectedIndex int
	searchQuery   string
	
	// UI components
	historyList   *widget.List
	searchEntry   *widget.Entry
	statusLabel   *widget.Label
	window        fyne.Window
	app           fyne.App
	
	// Control channels
	shutdownChan  chan struct{}
	running       bool
}

// NewClipboardManager creates a new clipboard manager instance
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
	
	// Load history from disk
	cm.loadHistory()
	cm.updateFiltered()
	
	return cm
}

// loadHistory loads clipboard history from file with error handling
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
		return
	}
}

// saveHistory saves clipboard history to file with error handling
func (cm *ClipboardManager) saveHistory() {
	cm.mu.RLock()
	historyCopy := make([]string, len(cm.history))
	copy(historyCopy, cm.history)
	cm.mu.RUnlock()
	
	// Save in background to avoid blocking UI
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

// updateFiltered updates the filtered list based on search query
func (cm *ClipboardManager) updateFiltered() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if cm.searchQuery == "" {
		cm.filtered = make([]string, len(cm.history))
		copy(cm.filtered, cm.history)
	} else {
		cm.filtered = cm.filtered[:0] // Reset slice but keep capacity
		query := strings.ToLower(cm.searchQuery)
		for _, item := range cm.history {
			if strings.Contains(strings.ToLower(item), query) {
				cm.filtered = append(cm.filtered, item)
			}
		}
	}
	cm.selectedIndex = -1
}

// addToHistory adds a new item to clipboard history
func (cm *ClipboardManager) addToHistory(content string) bool {
	if content == "" || len(content) > 10000 { // Ignore empty or very large content
		return false
	}
	
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Check if content already exists and is the last item
	if len(cm.history) > 0 && cm.history[len(cm.history)-1] == content {
		return false
	}
	
	// Remove duplicate if exists elsewhere
	for i := len(cm.history) - 1; i >= 0; i-- {
		if cm.history[i] == content {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}
	
	// Add to end and limit history size
	cm.history = append(cm.history, content)
	if len(cm.history) > 1000 { // Limit history to 1000 items
		cm.history = cm.history[1:]
	}
	
	return true
}

// deleteSelected removes the selected item from history
func (cm *ClipboardManager) deleteSelected() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if cm.selectedIndex < 0 || cm.selectedIndex >= len(cm.filtered) {
		return false
	}
	
	targetItem := cm.filtered[cm.selectedIndex]
	
	// Find and remove from main history
	for i, item := range cm.history {
		if item == targetItem {
			cm.history = append(cm.history[:i], cm.history[i+1:]...)
			break
		}
	}
	
	cm.selectedIndex = -1
	return true
}

// clearHistory clears all clipboard history
func (cm *ClipboardManager) clearHistory() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.history = []string{}
	cm.filtered = []string{}
	cm.selectedIndex = -1
}

// getFilteredCount returns the count of filtered items
func (cm *ClipboardManager) getFilteredCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.filtered)
}

// getFilteredItem returns a filtered item by index
func (cm *ClipboardManager) getFilteredItem(index int) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if index < 0 || index >= len(cm.filtered) {
		return ""
	}
	
	// Truncate long items for display
	item := cm.filtered[index]
	if len(item) > 100 {
		return item[:97] + "..."
	}
	return strings.ReplaceAll(item, "\n", " ") // Replace newlines with spaces
}

// getSelectedItem returns the currently selected item
func (cm *ClipboardManager) getSelectedItem() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if cm.selectedIndex < 0 || cm.selectedIndex >= len(cm.filtered) {
		return ""
	}
	return cm.filtered[cm.selectedIndex]
}

// setSelectedIndex sets the selected index
func (cm *ClipboardManager) setSelectedIndex(index int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.selectedIndex = index
}

// setSearchQuery sets the search query and updates filtered results
func (cm *ClipboardManager) setSearchQuery(query string) {
	cm.mu.Lock()
	cm.searchQuery = query
	cm.mu.Unlock()
	
	cm.updateFiltered()
}

// refreshUI safely refreshes UI components
func (cm *ClipboardManager) refreshUI() {
	if !cm.running {
		return
	}
	
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure UI thread is ready
		if cm.historyList != nil && cm.running {
			cm.historyList.Refresh()
		}
		if cm.statusLabel != nil && cm.running {
			count := cm.getFilteredCount()
			cm.statusLabel.SetText(fmt.Sprintf("Items: %d", count))
		}
	}()
}

// shutdown gracefully shuts down the clipboard manager
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

// startClipboardMonitor starts monitoring clipboard changes
func (cm *ClipboardManager) startClipboardMonitor() {
	go func() {
		var lastContent string
		ticker := time.NewTicker(500 * time.Millisecond) // Check every 500ms
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

// loadIcon loads the icon from file
func loadIcon() fyne.Resource {
	// Try to load icon.png from current directory
	iconPath := "icon.png"
	if data, err := os.ReadFile(iconPath); err == nil {
		return fyne.NewStaticResource("icon.png", data)
	}
	
	// If file doesn't exist, return nil (will use default icon)
	log.Printf("Icon file 'icon.png' not found in current directory")
	return nil
}

func main() {
	myApp := app.New()
	myApp.SetIcon(loadIcon()) // Load custom icon

	myWindow := myApp.NewWindow("Clipboard Manager")
	myWindow.Resize(fyne.NewSize(600, 500))

	// Create clipboard manager
	cm := NewClipboardManager(myWindow, myApp)

	// Create UI components
	historyList := widget.NewList(
		func() int {
			return cm.getFilteredCount()
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			item := cm.getFilteredItem(i)
			label.SetText(item)
		},
	)
	
	historyList.OnSelected = func(id widget.ListItemID) {
		cm.setSelectedIndex(id)
	}

	cm.historyList = historyList

	// Search entry
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search clipboard history...")
	searchEntry.OnChanged = func(query string) {
		cm.setSearchQuery(query)
		cm.refreshUI()
		historyList.UnselectAll()
	}

	cm.searchEntry = searchEntry

	// Buttons
	copyBtn := widget.NewButton("Copy Selected", func() {
		if selectedItem := cm.getSelectedItem(); selectedItem != "" {
			myWindow.Clipboard().SetContent(selectedItem)
		}
	})

	deleteBtn := widget.NewButton("Delete Selected", func() {
		if cm.deleteSelected() {
			cm.saveHistory()
			cm.updateFiltered()
			cm.refreshUI()
			historyList.UnselectAll()
		}
	})

	clearBtn := widget.NewButton("Clear All", func() {
		cm.clearHistory()
		cm.saveHistory()
		cm.refreshUI()
		historyList.UnselectAll()
	})

	// Status label
	statusLabel := widget.NewLabel(fmt.Sprintf("Items: %d", cm.getFilteredCount()))
	cm.statusLabel = statusLabel

	// Layout
	buttonContainer := container.NewHBox(
		copyBtn,
		deleteBtn,
		clearBtn,
	)

	content := container.NewBorder(
		searchEntry,        // top
		container.NewBorder(nil, nil, statusLabel, nil, buttonContainer), // bottom
		nil,               // left
		nil,               // right
		historyList,       // center
	)

	myWindow.SetContent(content)
	
	// Start clipboard monitoring
	cm.startClipboardMonitor()
	
	// Handle window close
	myWindow.SetCloseIntercept(func() {
		cm.shutdown()
		myWindow.Close()
	})

	myWindow.ShowAndRun()
}