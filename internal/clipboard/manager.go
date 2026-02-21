// File: internal/clipboard/manager.go
package clipboard

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	MaxHistoryItems = 1000
	UpdateDebounce  = 50 * time.Millisecond
	SaveDebounce    = 250 * time.Millisecond
)

// Manager handles clipboard history and operations
type Manager struct {
	mu       sync.RWMutex
	history  []Item
	filtered []Item
	storage  *Storage
	native   *NativeClipboard
	monitor  *Monitor

	selectedIndex int
	searchQuery   string
	lastCopied    time.Time

	updateChan   chan struct{}
	saveChan     chan struct{}
	shutdownChan chan struct{}
	running      bool

	// Callbacks
	onUpdate func()
	onError  func(error)
}

// Config holds manager configuration
type Config struct {
	StoragePath string
	OnUpdate    func()
	OnError     func(error)
}

// NewManager creates a new clipboard manager
func NewManager(cfg Config) (*Manager, error) {
	storage, err := NewStorage(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	native, err := NewNativeClipboard()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	m := &Manager{
		storage:       storage,
		native:        native,
		selectedIndex: -1,
		updateChan:    make(chan struct{}, 100),
		saveChan:      make(chan struct{}, 1),
		shutdownChan:  make(chan struct{}),
		running:       true,
		onUpdate:      cfg.OnUpdate,
		onError:       cfg.OnError,
	}

	// Load history
	if err := m.loadHistory(); err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	// Start update dispatcher
	go m.updateDispatcher()
	go m.saveDispatcher()

	// Create and start monitor
	m.monitor = NewMonitor(m, native)
	m.monitor.Start()

	return m, nil
}

// loadHistory loads history from storage
func (m *Manager) loadHistory() error {
	items, err := m.storage.Load()
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.history = items
	m.mu.Unlock()

	m.updateFiltered()
	return nil
}

// saveHistory persists history to storage (internal)
func (m *Manager) saveHistory() {
	select {
	case m.saveChan <- struct{}{}:
	default:
		// Save already queued
	}
}

// persistHistory writes a snapshot to storage.
func (m *Manager) persistHistory() {
	m.mu.RLock()
	items := make([]Item, len(m.history))
	copy(items, m.history)
	m.mu.RUnlock()

	if err := m.storage.Save(items); err != nil {
		if m.onError != nil {
			m.onError(fmt.Errorf("failed to save history: %w", err))
		}
	}
}

// SaveHistory is a public method to force save history
func (m *Manager) SaveHistory() {
	m.saveHistory()
}

// AddItem adds a new item to history
func (m *Manager) AddItem(item Item) bool {
	if item.Content == "" && item.ImageData == "" {
		return false
	}

	// Generate hash if not provided
	if item.Hash == "" {
		if item.Type == TypeText {
			hash := sha256.Sum256([]byte(item.Content))
			item.Hash = hex.EncodeToString(hash[:])
		} else {
			hash := sha256.Sum256([]byte(item.ImageData))
			item.Hash = hex.EncodeToString(hash[:])
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate at end
	if len(m.history) > 0 {
		last := m.history[len(m.history)-1]
		if last.Hash == item.Hash && last.Type == item.Type {
			return false
		}
	}

	// Remove existing duplicates
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].Hash == item.Hash && m.history[i].Type == item.Type {
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}

	// Add new item
	item.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	item.Timestamp = time.Now()
	m.history = append(m.history, item)

	// Trim unpinned items
	m.trimHistory()
	m.lastCopied = time.Now()

	return true
}

// trimHistory keeps only the last MaxHistoryItems unpinned items
func (m *Manager) trimHistory() {
	unpinnedCount := 0
	for i := len(m.history) - 1; i >= 0; i-- {
		if !m.history[i].Pinned {
			unpinnedCount++
			if unpinnedCount > MaxHistoryItems {
				m.history = append(m.history[:i], m.history[i+1:]...)
			}
		}
	}
}

// updateFiltered updates the filtered list based on search
func (m *Manager) updateFiltered() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Separate pinned and unpinned
	var pinned, unpinned []Item
	for _, item := range m.history {
		if item.Pinned {
			pinned = append(pinned, item)
		} else {
			unpinned = append(unpinned, item)
		}
	}

	// Combine with pinned first
	sorted := append(pinned, unpinned...)

	// Apply search filter
	if m.searchQuery == "" {
		m.filtered = sorted
	} else {
		m.filtered = m.filtered[:0]
		query := strings.ToLower(m.searchQuery)
		for _, item := range sorted {
			if strings.Contains(strings.ToLower(item.Content), query) {
				m.filtered = append(m.filtered, item)
			}
		}
	}

	m.selectedIndex = -1
}

// SetSearch updates the search query
func (m *Manager) SetSearch(query string) {
	m.mu.Lock()
	m.searchQuery = query
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
}

// GetFiltered returns the filtered items (thread-safe)
func (m *Manager) GetFiltered() []Item {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]Item, len(m.filtered))
	copy(items, m.filtered)
	return items
}

// GetFilteredCount returns the count of filtered items
func (m *Manager) GetFilteredCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.filtered)
}

// GetItem returns an item by index (thread-safe)
func (m *Manager) GetItem(index int) (Item, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if index < 0 || index >= len(m.filtered) {
		return Item{}, false
	}
	return m.filtered[index], true
}

// GetSelected returns the currently selected item
func (m *Manager) GetSelected() (Item, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.selectedIndex < 0 || m.selectedIndex >= len(m.filtered) {
		return Item{}, false
	}
	return m.filtered[m.selectedIndex], true
}

// GetSelectedIndex returns the selected index
func (m *Manager) GetSelectedIndex() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.selectedIndex
}

// SetSelected sets the selected index
func (m *Manager) SetSelected(index int) {
	m.mu.Lock()
	m.selectedIndex = index
	m.mu.Unlock()
}

// TogglePin toggles the pin status of an item
func (m *Manager) TogglePin(index int) bool {
	m.mu.Lock()

	if index < 0 || index >= len(m.filtered) {
		m.mu.Unlock()
		return false
	}

	targetID := m.filtered[index].ID
	found := false
	for i := range m.history {
		if m.history[i].ID == targetID {
			m.history[i].Pinned = !m.history[i].Pinned
			found = true
			break
		}
	}

	m.mu.Unlock()

	if found {
		m.updateFiltered()
		m.triggerUpdate()
	}

	return found
}

// Delete removes an item
func (m *Manager) Delete(index int) error {
	m.mu.Lock()

	if index < 0 || index >= len(m.filtered) {
		m.mu.Unlock()
		return fmt.Errorf("invalid index")
	}

	targetItem := m.filtered[index]
	if targetItem.Pinned {
		m.mu.Unlock()
		return fmt.Errorf("cannot delete pinned items")
	}

	found := false
	for i, item := range m.history {
		if item.ID == targetItem.ID {
			m.history = append(m.history[:i], m.history[i+1:]...)
			found = true
			break
		}
	}

	m.selectedIndex = -1
	m.mu.Unlock()

	if found {
		m.updateFiltered()
		m.triggerUpdate()
	} else {
		return fmt.Errorf("item not found")
	}

	return nil
}

// ClearUnpinned removes all unpinned items
func (m *Manager) ClearUnpinned() {
	m.mu.Lock()

	var pinned []Item
	for _, item := range m.history {
		if item.Pinned {
			pinned = append(pinned, item)
		}
	}

	m.history = pinned
	m.selectedIndex = -1
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
}

// CopyToClipboard copies an item to the system clipboard
func (m *Manager) CopyToClipboard(index int) error {
	item, ok := m.GetItem(index)
	if !ok {
		return fmt.Errorf("invalid index")
	}

	// Notify monitor about programmatic copy
	if m.monitor != nil {
		if item.Type == TypeText {
			m.monitor.SetProgrammaticCopy([]byte(item.Content))
		} else {
			m.monitor.SetProgrammaticCopy([]byte(item.ImageData))
		}
	}

	m.mu.Lock()
	m.lastCopied = time.Now()
	m.mu.Unlock()

	if item.Type == TypeText {
		return m.native.WriteText([]byte(item.Content))
	}

	return m.native.WriteImage(item.ImageData)
}

// GetStats returns statistics about the clipboard
func (m *Manager) GetStats() (total, pinned, filtered int, lastCopied time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total = len(m.history)
	filtered = len(m.filtered)
	lastCopied = m.lastCopied

	for _, item := range m.history {
		if item.Pinned {
			pinned++
		}
	}

	return
}

// triggerUpdate signals that the UI should be updated
func (m *Manager) triggerUpdate() {
	if !m.running {
		return
	}

	select {
	case m.updateChan <- struct{}{}:
	default:
		// Channel full, update already queued
	}
}

// updateDispatcher handles debounced UI updates
func (m *Manager) updateDispatcher() {
	var timer *time.Timer

	for {
		select {
		case <-m.updateChan:
			// Drain channel
			for len(m.updateChan) > 0 {
				<-m.updateChan
			}

			// Reset debounce timer
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(UpdateDebounce, func() {
				if m.running && m.onUpdate != nil {
					m.onUpdate()
				}
			})

		case <-m.shutdownChan:
			if timer != nil {
				timer.Stop()
			}
			return
		}
	}
}

// saveDispatcher debounces and serializes history saves.
func (m *Manager) saveDispatcher() {
	var timer *time.Timer
	var timerCh <-chan time.Time

	for {
		select {
		case <-m.saveChan:
			if timer == nil {
				timer = time.NewTimer(SaveDebounce)
				timerCh = timer.C
				continue
			}

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(SaveDebounce)
			timerCh = timer.C

		case <-timerCh:
			m.persistHistory()
			timerCh = nil

		case <-m.shutdownChan:
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
			return
		}
	}
}

// Shutdown stops the manager
func (m *Manager) Shutdown() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	if m.monitor != nil {
		m.monitor.Stop()
	}

	select {
	case <-m.shutdownChan:
	default:
		close(m.shutdownChan)
	}

	m.persistHistory()
}
