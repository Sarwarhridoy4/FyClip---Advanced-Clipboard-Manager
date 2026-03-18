// File: internal/clipboard/manager.go
package clipboard

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MaxHistoryItems = 1000
	UpdateDebounce  = 200 * time.Millisecond
	SaveDebounce    = 250 * time.Millisecond
)

// AddItemResult describes the result of adding a clipboard item.
type AddItemResult struct {
	Added          bool
	MovedDuplicate bool
}

// Manager handles clipboard history and operations
type Manager struct {
	mu       sync.RWMutex
	history  []Item
	filtered []Item
	storage  *Storage
	native   *NativeClipboard
	monitor  *Monitor

	// Index maps for O(1) lookups
	hashIndexMap map[string]int  // hash+type -> index in history
	idIndexMap   map[string]int  // id -> index in history

	selectedIndex   int
	searchQuery     string
	showPinnedOnly  bool
	lastCopied      time.Time
	maxHistoryItems int

	updateChan   chan struct{}
	saveChan     chan struct{}
	shutdownChan chan struct{}
	running      bool

	// Callbacks
	onUpdate func()
	onError  func(error)
	onInfo   func(string)
}

// Config holds manager configuration
type Config struct {
	StoragePath string
	OnUpdate    func()
	OnError     func(error)
	OnInfo      func(string)
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
		storage:         storage,
		native:          native,
		selectedIndex:   -1,
		updateChan:      make(chan struct{}, 100),
		saveChan:        make(chan struct{}, 1),
		shutdownChan:    make(chan struct{}),
		running:         true,
		maxHistoryItems: MaxHistoryItems,
		onUpdate:        cfg.OnUpdate,
		onError:         cfg.OnError,
		onInfo:          cfg.OnInfo,
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
	m.buildIndexMaps()
	m.mu.Unlock()

	m.updateFiltered()
	return nil
}

// buildIndexMaps rebuilds the hash and ID index maps from history
func (m *Manager) buildIndexMaps() {
	m.hashIndexMap = make(map[string]int, len(m.history))
	m.idIndexMap = make(map[string]int, len(m.history))
	for i, item := range m.history {
		// Create composite key with type for hash to handle same content different types
		hashKey := item.Hash + ":" + strconv.Itoa(int(item.Type))
		m.hashIndexMap[hashKey] = i
		m.idIndexMap[item.ID] = i
	}
}

// queueSaveHistory schedules a debounced save of history.
func (m *Manager) queueSaveHistory() {
	if !m.running {
		return
	}

	select {
	case m.saveChan <- struct{}{}:
	default:
		// Save already queued.
	}
}

// saveHistory persists history to storage (internal)
func (m *Manager) saveHistory() {
	m.queueSaveHistory()
}

// saveHistoryNow writes a snapshot to storage immediately.
func (m *Manager) saveHistoryNow() {
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
func (m *Manager) AddItem(item Item) AddItemResult {
	if item.Content == "" && item.ImageData == "" {
		return AddItemResult{}
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

	// Create composite key with type for hash
	hashKey := item.Hash + ":" + strconv.Itoa(int(item.Type))

	// Check for duplicate at end (last item check)
	if len(m.history) > 0 {
		last := m.history[len(m.history)-1]
		lastHashKey := last.Hash + ":" + strconv.Itoa(int(last.Type))
		if lastHashKey == hashKey {
			return AddItemResult{}
		}
	}

	// O(1) lookup for existing duplicate using hashIndexMap
	movedDuplicate := false
	if idx, exists := m.hashIndexMap[hashKey]; exists {
		// Remove the existing item
		m.removeAtIndex(idx)
		movedDuplicate = true
	}

	// Add new item
	item.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	item.Timestamp = time.Now()

	// Add to history and update index maps
	newIndex := len(m.history)
	m.history = append(m.history, item)
	m.hashIndexMap[hashKey] = newIndex
	m.idIndexMap[item.ID] = newIndex

	// Trim unpinned items
	m.trimHistory()
	m.lastCopied = time.Now()

	return AddItemResult{
		Added:          true,
		MovedDuplicate: movedDuplicate,
	}
}

// removeAtIndex removes an item at the given index and updates index maps
// Caller must hold the lock
func (m *Manager) removeAtIndex(idx int) {
	if idx < 0 || idx >= len(m.history) {
		return
	}

	item := m.history[idx]
	hashKey := item.Hash + ":" + strconv.Itoa(int(item.Type))

	// Remove from index maps
	delete(m.hashIndexMap, hashKey)
	delete(m.idIndexMap, item.ID)

	// Remove from history
	m.history = append(m.history[:idx], m.history[idx+1:]...)

	// Rebuild index maps for indices after removed position
	// This is necessary because all subsequent indices shift by 1
	m.rebuildIndexMapsFrom(idx)
}

// rebuildIndexMapsFrom rebuilds index maps starting from the given index
// Caller must hold the lock
func (m *Manager) rebuildIndexMapsFrom(startIdx int) {
	for i := startIdx; i < len(m.history); i++ {
		item := m.history[i]
		hashKey := item.Hash + ":" + strconv.Itoa(int(item.Type))
		m.hashIndexMap[hashKey] = i
		m.idIndexMap[item.ID] = i
	}
}

// trimHistory keeps only the last MaxHistoryItems unpinned items
func (m *Manager) trimHistory() {
	unpinnedCount := 0
	for i := len(m.history) - 1; i >= 0; i-- {
		if !m.history[i].Pinned {
			unpinnedCount++
			if unpinnedCount > m.maxHistoryItems {
				m.removeAtIndex(i)
			}
		}
	}
}

// updateFiltered updates the filtered list based on search
func (m *Manager) updateFiltered() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cap(m.filtered) < len(m.history) {
		m.filtered = make([]Item, 0, len(m.history))
	} else {
		m.filtered = m.filtered[:0]
	}

	query := strings.ToLower(m.searchQuery)
	match := func(item *Item) bool {
		if query == "" {
			return true
		}
		if item.SearchContent() == "" && item.Content != "" {
			item.PrepareForSearch()
		}
		return strings.Contains(item.SearchContent(), query)
	}

	// Keep pinned items first while preserving stable order in each group.
	for i := range m.history {
		item := &m.history[i]
		if !item.Pinned {
			continue
		}
		if match(item) {
			m.filtered = append(m.filtered, *item)
		}
	}
	for i := range m.history {
		item := &m.history[i]
		if item.Pinned {
			continue
		}
		if m.showPinnedOnly {
			continue
		}
		if match(item) {
			m.filtered = append(m.filtered, *item)
		}
	}

	m.selectedIndex = -1
}

// SetPinnedOnly controls whether only pinned items are shown.
func (m *Manager) SetPinnedOnly(enabled bool) {
	m.mu.Lock()
	m.showPinnedOnly = enabled
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
}

// TogglePinnedOnly toggles pinned-only filtering and returns the new state.
func (m *Manager) TogglePinnedOnly() bool {
	m.mu.Lock()
	m.showPinnedOnly = !m.showPinnedOnly
	enabled := m.showPinnedOnly
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
	return enabled
}

// IsPinnedOnly returns whether pinned-only filter is enabled.
func (m *Manager) IsPinnedOnly() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.showPinnedOnly
}

// SetMaxHistory updates the max number of unpinned items retained.
func (m *Manager) SetMaxHistory(limit int) bool {
	if limit <= 0 {
		return false
	}

	m.mu.Lock()
	m.maxHistoryItems = limit
	m.trimHistory()
	m.mu.Unlock()

	m.updateFiltered()
	m.saveHistory()
	m.triggerUpdate()
	return true
}

// GetMaxHistory returns the configured max unpinned history count.
func (m *Manager) GetMaxHistory() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxHistoryItems
}

// PauseMonitoringFor pauses clipboard capture for the specified duration.
func (m *Manager) PauseMonitoringFor(d time.Duration) {
	if m.monitor == nil {
		return
	}
	m.monitor.PauseFor(d)
}

// ResumeMonitoring resumes clipboard capture immediately.
func (m *Manager) ResumeMonitoring() {
	if m.monitor == nil {
		return
	}
	m.monitor.Resume()
}

// IsMonitoringPaused reports whether capture is paused.
func (m *Manager) IsMonitoringPaused() bool {
	if m.monitor == nil {
		return false
	}
	return m.monitor.IsPaused()
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

	// O(1) lookup using idIndexMap
	idx, exists := m.idIndexMap[targetID]
	if !exists {
		m.mu.Unlock()
		return false
	}

	m.history[idx].Pinned = !m.history[idx].Pinned

	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()

	return true
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

	// O(1) lookup using idIndexMap
	idx, exists := m.idIndexMap[targetItem.ID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("item not found")
	}

	m.removeAtIndex(idx)

	m.selectedIndex = -1
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()

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
	m.buildIndexMaps()
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
			if rawImage, err := base64.StdEncoding.DecodeString(item.ImageData); err == nil {
				m.monitor.SetProgrammaticCopy(rawImage)
			} else {
				m.monitor.SetProgrammaticCopy([]byte(item.ImageData))
			}
		}
	}

	m.mu.Lock()
	m.lastCopied = time.Now()
	m.mu.Unlock()

	if item.Type == TypeText {
		if err := m.native.WriteText([]byte(item.Content)); err != nil {
			return err
		}
	} else {
		if err := m.native.WriteImage(item.ImageData); err != nil {
			return err
		}
	}

	m.incrementCopyCount(item.ID)
	m.saveHistory()
	m.triggerUpdate()
	return nil
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

// saveDispatcher coalesces frequent save requests into fewer disk writes.
func (m *Manager) saveDispatcher() {
	var (
		timer  *time.Timer
		timerC <-chan time.Time
	)

	resetTimer := func() {
		if timer == nil {
			timer = time.NewTimer(SaveDebounce)
			timerC = timer.C
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(SaveDebounce)
		timerC = timer.C
	}

	for {
		select {
		case <-m.saveChan:
			for len(m.saveChan) > 0 {
				<-m.saveChan
			}
			resetTimer()

		case <-timerC:
			m.saveHistoryNow()
			timerC = nil

		case <-m.shutdownChan:
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
			flush := false
			select {
			case <-m.saveChan:
				flush = true
			default:
			}
			for len(m.saveChan) > 0 {
				<-m.saveChan
				flush = true
			}
			if flush {
				m.saveHistoryNow()
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

	m.saveHistoryNow()
}

func (m *Manager) incrementCopyCount(itemID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// O(1) lookup using idIndexMap
	if idx, exists := m.idIndexMap[itemID]; exists {
		m.history[idx].CopyCount++
	}
}

func (m *Manager) notifyInfo(message string) {
	if m.onInfo != nil {
		m.onInfo(message)
	}
}
