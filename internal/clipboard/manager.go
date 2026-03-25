// File: internal/clipboard/manager.go
package clipboard

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MaxHistoryItems = 1000
	UpdateDebounce  = 200 * time.Millisecond
	SaveDebounce    = 250 * time.Millisecond

	// Shutdown timeout for graceful shutdown
	ShutdownTimeout = 10 * time.Second
)

// AddItemResult describes the result of adding a clipboard item.
type AddItemResult struct {
	Added          bool
	MovedDuplicate bool
}

// ShutdownHook is a function that runs during shutdown
type ShutdownHook func(ctx context.Context) error

// Manager handles clipboard history and operations
type Manager struct {
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	history  []Item
	filtered []*Item  // Changed to pointers to reduce memory copies
	storage  *Storage
	snippets  *SnippetManager
	exclusions *ExclusionManager
	native   *NativeClipboard
	monitor  *Monitor
	backup   *BackupManager

	// Shutdown hooks
	shutdownHooks []ShutdownHook

	// Index maps for O(1) lookups
	hashIndexMap map[string]int  // hash+type -> index in history
	idIndexMap   map[string]int  // id -> index in history

	selectedIndex   int
	searchQuery     string
	searchOptions   *SearchOptions
	showPinnedOnly  bool
	lastCopied      time.Time
	maxHistoryItems int

	updateChan   chan struct{}
	saveChan     chan struct{}
	shutdownChan chan struct{}
	running      bool

	// Memory tracking
	lastMemoryCheck time.Time

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
	ctx, cancel := context.WithCancel(context.Background())

	storage, err := NewStorage(cfg.StoragePath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	native, err := NewNativeClipboard()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	m := &Manager{
		ctx:             ctx,
		cancel:          cancel,
		storage:         storage,
		snippets:       NewSnippetManager(storage),
		exclusions:    NewExclusionManager(),
		native:        native,
		backup:        NewBackupManager(storage),
		hashIndexMap:  make(map[string]int),
		idIndexMap:    make(map[string]int),
		selectedIndex: -1,
		updateChan:     make(chan struct{}, 10), // Reduced from 100 - updates are debounced
		saveChan:       make(chan struct{}, 1),
		shutdownChan:  make(chan struct{}),
		running:       true,
		maxHistoryItems: MaxHistoryItems,
		searchOptions:  DefaultSearchOptions(),
		onUpdate:       cfg.OnUpdate,
		onError:        cfg.OnError,
		onInfo:         cfg.OnInfo,
	}

	// Load snippets
	if err := m.snippets.LoadSnippets(); err != nil {
		// Log but don't fail - snippets are optional
		if cfg.OnError != nil {
			cfg.OnError(fmt.Errorf("failed to load snippets: %w", err))
		}
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

// ReloadHistory reloads history from storage
func (m *Manager) ReloadHistory() error {
	return m.loadHistory()
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
	newItem := GetFromPool()
	*newItem = item
	newItem.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	newItem.Timestamp = time.Now()

	// Auto-detect category based on content
	newItem.AutoDetectCategory()

	// Add to history and update index maps
	newIndex := len(m.history)
	m.history = append(m.history, *newItem)
	m.hashIndexMap[hashKey] = newIndex
	m.idIndexMap[newItem.ID] = newIndex

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

	// Return removed item to pool
	ReturnToPool(&item)

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
		m.filtered = make([]*Item, 0, len(m.history))
	} else {
		m.filtered = m.filtered[:0]
	}

	query := m.searchQuery
	searchOpts := m.searchOptions

	match := func(item *Item) bool {
		if query == "" {
			return true
		}
		return SearchItem(item, query, searchOpts)
	}

	// Keep pinned items first while preserving stable order in each group.
	for i := range m.history {
		item := &m.history[i]
		if !item.Pinned {
			continue
		}
		if match(item) {
			m.filtered = append(m.filtered, item)
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
			m.filtered = append(m.filtered, item)
		}
	}

	m.selectedIndex = -1
}

// CheckMemoryPressure checks memory usage and triggers cleanup if needed
func (m *Manager) CheckMemoryPressure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	// Only check every 30 seconds to avoid overhead
	if now.Sub(m.lastMemoryCheck) < 30*time.Second {
		return
	}
	m.lastMemoryCheck = now

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// If heap usage > 100MB, reduce history size by 20%
	if memStats.Alloc > 100*1024*1024 {
		oldMax := m.maxHistoryItems
		m.maxHistoryItems = int(float64(m.maxHistoryItems) * 0.8)
		if m.maxHistoryItems < 100 {
			m.maxHistoryItems = 100
		}
		if m.maxHistoryItems < oldMax {
			m.trimHistory()
			if m.onInfo != nil {
				m.onInfo(fmt.Sprintf("Memory pressure: reduced history from %d to %d", oldMax, m.maxHistoryItems))
			}
		}
	}
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

// GetBackupManager returns the backup manager
func (m *Manager) GetBackupManager() *BackupManager {
	return m.backup
}

// SetSearch updates the search query
func (m *Manager) SetSearch(query string) {
	m.mu.Lock()
	m.searchQuery = query
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
}

// SetSearchOptions updates the search options
func (m *Manager) SetSearchOptions(opts *SearchOptions) {
	m.mu.Lock()
	m.searchOptions = opts
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
}

// GetSearchOptions returns the current search options
func (m *Manager) GetSearchOptions() *SearchOptions {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.searchOptions
}

// ToggleRegexSearch toggles regex search mode
func (m *Manager) ToggleRegexSearch() bool {
	m.mu.Lock()
	m.searchOptions.RegexEnabled = !m.searchOptions.RegexEnabled
	enabled := m.searchOptions.RegexEnabled
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
	return enabled
}

// ToggleCaseSensitive toggles case-sensitive search
func (m *Manager) ToggleCaseSensitive() bool {
	m.mu.Lock()
	m.searchOptions.CaseSensitive = !m.searchOptions.CaseSensitive
	enabled := m.searchOptions.CaseSensitive
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
	return enabled
}

// ToggleFuzzySearch toggles fuzzy search mode
func (m *Manager) ToggleFuzzySearch() bool {
	m.mu.Lock()
	m.searchOptions.FuzzyEnabled = !m.searchOptions.FuzzyEnabled
	enabled := m.searchOptions.FuzzyEnabled
	m.mu.Unlock()

	m.updateFiltered()
	m.triggerUpdate()
	return enabled
}

// GetFiltered returns the filtered items (thread-safe)
func (m *Manager) GetFiltered() []Item {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]Item, len(m.filtered))
	for i, p := range m.filtered {
		items[i] = *p
	}
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
	return *m.filtered[index], true
}

// GetSelected returns the currently selected item
func (m *Manager) GetSelected() (Item, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.selectedIndex < 0 || m.selectedIndex >= len(m.filtered) {
		return Item{}, false
	}
	return *m.filtered[m.selectedIndex], true
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

// FindIndexByID finds the current index of an item by its ID in the filtered list
func (m *Manager) FindIndexByID(id string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i, item := range m.filtered {
		if item.ID == id {
		return i
		}
	}
	return -1
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

// BulkDelete removes multiple items by their IDs
// Returns the count of successfully deleted items
func (m *Manager) BulkDelete(ids []string) int {
	m.mu.Lock()
	deleted := 0

	for _, id := range ids {
		// O(1) lookup using idIndexMap
		idx, exists := m.idIndexMap[id]
		if !exists {
			continue
		}

		// Skip pinned items
		if m.history[idx].Pinned {
			continue
		}

		m.removeAtIndex(idx)
		deleted++
	}

	m.selectedIndex = -1
	m.mu.Unlock()

	if deleted > 0 {
		m.updateFiltered()
		m.triggerUpdate()
	}

	return deleted
}

// BulkPin pins multiple items by their IDs
// Returns the count of successfully pinned items
func (m *Manager) BulkPin(ids []string) int {
	m.mu.Lock()
	pinned := 0

	for _, id := range ids {
		// O(1) lookup using idIndexMap
		idx, exists := m.idIndexMap[id]
		if !exists {
			continue
		}

		if !m.history[idx].Pinned {
			m.history[idx].Pinned = true
			pinned++
		}
	}

	m.mu.Unlock()

	if pinned > 0 {
		m.updateFiltered()
		m.triggerUpdate()
	}

	return pinned
}

// BulkUnpin unpins multiple items by their IDs
// Returns the count of successfully unpinned items
func (m *Manager) BulkUnpin(ids []string) int {
	m.mu.Lock()
	unpinned := 0

	for _, id := range ids {
		// O(1) lookup using idIndexMap
		idx, exists := m.idIndexMap[id]
		if !exists {
			continue
		}

		if m.history[idx].Pinned {
			m.history[idx].Pinned = false
			unpinned++
		}
	}

	m.mu.Unlock()

	if unpinned > 0 {
		m.updateFiltered()
		m.triggerUpdate()
	}

	return unpinned
}

// BulkTogglePin toggles pin status for multiple items
// Returns the count of items that were toggled
func (m *Manager) BulkTogglePin(ids []string) int {
	m.mu.Lock()
	toggled := 0

	for _, id := range ids {
		// O(1) lookup using idIndexMap
		idx, exists := m.idIndexMap[id]
		if !exists {
			continue
		}

		m.history[idx].Pinned = !m.history[idx].Pinned
		toggled++
	}

	m.mu.Unlock()

	if toggled > 0 {
		m.updateFiltered()
		m.triggerUpdate()
	}

	return toggled
}

// BulkCopy copies multiple items to clipboard sequentially
// Returns the count of successfully copied items
func (m *Manager) BulkCopy(indices []int) int {
	copied := 0

	for _, index := range indices {
		if err := m.CopyToClipboard(index); err == nil {
			copied++
		}
	}

	return copied
}

// CopyToClipboard copies an item to the system clipboard
func (m *Manager) CopyToClipboard(index int) error {
	item, ok := m.GetItem(index)
	if !ok {
		return fmt.Errorf("invalid index")
	}

	return m.copyItemToClipboard(&item, false)
}

// CopyToClipboardAsHTML copies an item to clipboard as HTML (if available)
func (m *Manager) CopyToClipboardAsHTML(index int) error {
	item, ok := m.GetItem(index)
	if !ok {
		return fmt.Errorf("invalid index")
	}

	return m.copyItemToClipboard(&item, true)
}

// CopyToClipboardAsPlainText copies an item to clipboard as plain text
func (m *Manager) CopyToClipboardAsPlainText(index int) error {
	item, ok := m.GetItem(index)
	if !ok {
		return fmt.Errorf("invalid index")
	}

	return m.copyItemToClipboard(&item, false)
}

// copyItemToClipboard is the internal method for copying items
func (m *Manager) copyItemToClipboard(item *Item, asHTML bool) error {
	// Notify monitor about programmatic copy
	if m.monitor != nil {
		if item.Type == TypeText || item.Type == TypeHTML {
			m.monitor.SetProgrammaticCopy([]byte(item.Content))
		} else if item.Type == TypeImage {
			if rawImage, err := base64.StdEncoding.DecodeString(item.ImageData); err == nil {
				m.monitor.SetProgrammaticCopy(rawImage)
			} else {
				m.monitor.SetProgrammaticCopy([]byte(item.ImageData))
			}
		} else if item.Type == TypeFile && item.FileInfo != nil {
			m.monitor.SetProgrammaticCopy([]byte(item.FileInfo.Path))
		}
	}

	m.mu.Lock()
	m.lastCopied = time.Now()
	m.mu.Unlock()

	switch item.Type {
	case TypeText:
		if err := m.native.WriteText([]byte(item.Content)); err != nil {
			return err
		}
	case TypeHTML:
		// If asHTML is true and we have HTML content, write as HTML
		if asHTML && item.HasHTML() {
			if err := m.native.WriteHTML(item.HTMLContent, item.Content); err != nil {
				return err
			}
		} else {
			// Write plain text
			if err := m.native.WriteText([]byte(item.Content)); err != nil {
				return err
			}
		}
	case TypeImage:
		if err := m.native.WriteImage(item.ImageData); err != nil {
			return err
		}
	case TypeFile:
		if item.FileInfo != nil {
			// Copy file path to clipboard
			if err := m.native.WriteText([]byte(item.FileInfo.Path)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported item type")
	}

	m.incrementCopyCount(item.ID)
	m.saveHistory()
	m.triggerUpdate()
	return nil
}

// OpenFileLocation opens the file location in the system file manager
func (m *Manager) OpenFileLocation(index int) error {
	item, ok := m.GetItem(index)
	if !ok {
		return fmt.Errorf("invalid index")
	}

	if !item.IsFile() || item.FileInfo == nil {
		return fmt.Errorf("item is not a file")
	}

	path := item.FileInfo.Path
	// Get directory path
	dirPath := path
	if !item.FileInfo.IsDirectory {
		// Get the directory containing the file
		lastSlash := strings.LastIndex(path, "/")
		if lastSlash > 0 {
			dirPath = path[:lastSlash]
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", dirPath)
	case "darwin":
		cmd = exec.Command("open", dirPath)
	case "windows":
		cmd = exec.Command("explorer", dirPath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
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

// Shutdown stops the manager gracefully with timeout
func (m *Manager) Shutdown() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	// Cancel context to stop all goroutines
	if m.cancel != nil {
		m.cancel()
	}

	// Stop the monitor
	if m.monitor != nil {
		m.monitor.Stop()
	}

	// Run shutdown hooks with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	m.runShutdownHooks(ctx)

	// Close channels
	close(m.shutdownChan)

	// Final save
	m.saveHistoryNow()
}

// runShutdownHooks executes all registered shutdown hooks
func (m *Manager) runShutdownHooks(ctx context.Context) {
	m.mu.RLock()
	hooks := make([]ShutdownHook, len(m.shutdownHooks))
	copy(hooks, m.shutdownHooks)
	m.mu.RUnlock()

	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := hook(ctx); err != nil {
			// Log error but continue with other hooks
			if m.onError != nil {
				m.onError(fmt.Errorf("shutdown hook error: %w", err))
			}
		}
	}
}

// AddShutdownHook adds a function to be called during shutdown
func (m *Manager) AddShutdownHook(hook ShutdownHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownHooks = append(m.shutdownHooks, hook)
}

// RemoveShutdownHook removes a shutdown hook by index
func (m *Manager) RemoveShutdownHook(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index >= 0 && index < len(m.shutdownHooks) {
		m.shutdownHooks = append(m.shutdownHooks[:index], m.shutdownHooks[index+1:]...)
	}
}

// ClearShutdownHooks removes all shutdown hooks
func (m *Manager) ClearShutdownHooks() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownHooks = nil
}

// GetContext returns the manager's context
func (m *Manager) GetContext() context.Context {
	return m.ctx
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

// GetSnippets returns all snippets
func (m *Manager) GetSnippets() []Snippet {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snippets.GetSnippets()
}

// GetSnippetByAbbreviation returns a snippet by its abbreviation
func (m *Manager) GetSnippetByAbbreviation(abbr string) (Snippet, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snippets.GetSnippetByAbbreviation(abbr)
}

// AddSnippet adds a new snippet
func (m *Manager) AddSnippet(snippet Snippet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snippets.AddSnippet(snippet)
}

// UpdateSnippet updates an existing snippet
func (m *Manager) UpdateSnippet(snippet Snippet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snippets.UpdateSnippet(snippet)
}

// DeleteSnippet deletes a snippet
func (m *Manager) DeleteSnippet(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snippets.DeleteSnippet(id)
}

// ExpandSnippet expands template variables in content
func (m *Manager) ExpandSnippet(content string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Get current clipboard content if available
	clipContent := ""
	if len(m.history) > 0 {
		clipContent = m.history[0].Content
	}
	return m.snippets.ExpandSnippet(content, clipContent)
}

// ExpandSnippetWithClipboard expands template variables with provided clipboard content
func (m *Manager) ExpandSnippetWithClipboard(content string, clipboardContent string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.snippets.ExpandSnippet(content, clipboardContent)
}

// GetExclusionRules returns all exclusion rules
func (m *Manager) GetExclusionRules() []ExclusionRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.exclusions.GetRules()
}

// AddExclusionRule adds a new exclusion rule
func (m *Manager) AddExclusionRule(rule ExclusionRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exclusions.AddRule(rule)
}

// RemoveExclusionRule removes an exclusion rule
func (m *Manager) RemoveExclusionRule(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exclusions.RemoveRule(id)
}

// UpdateExclusionRule updates an exclusion rule
func (m *Manager) UpdateExclusionRule(rule ExclusionRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exclusions.UpdateRule(rule)
}

// ShouldExcludeContent checks if content should be excluded
func (m *Manager) ShouldExcludeContent(content string, contentSize int, appName string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.exclusions.ShouldExclude(content, contentSize, appName)
}

// DeleteMultiple deletes multiple items by their indices in the filtered list
// Returns the number of items successfully deleted
func (m *Manager) DeleteMultiple(indices []int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(indices) == 0 {
		return 0
	}

	// Sort indices in reverse order to delete from end to start
	// This prevents index shifting issues
	sortedIndices := make([]int, len(indices))
	copy(sortedIndices, indices)
	for i := 0; i < len(sortedIndices)-1; i++ {
		for j := i + 1; j < len(sortedIndices); j++ {
			if sortedIndices[i] > sortedIndices[j] {
				sortedIndices[i], sortedIndices[j] = sortedIndices[j], sortedIndices[i]
			}
		}
	}

	deleted := 0
	seen := make(map[int]bool)
	for _, idx := range sortedIndices {
		// Skip duplicates
		if seen[idx] {
			continue
		}
		seen[idx] = true

		if idx < 0 || idx >= len(m.filtered) {
			continue
		}

		targetItem := m.filtered[idx]
		if targetItem.Pinned {
			continue
		}

		// O(1) lookup using idIndexMap
		histIdx, exists := m.idIndexMap[targetItem.ID]
		if !exists {
			continue
		}

		m.removeAtIndex(histIdx)
		deleted++
	}

	m.selectedIndex = -1

	if deleted > 0 {
		m.queueSaveHistory()
	}

	return deleted
}

// TogglePinMultiple toggles the pin status of multiple items
// Returns the number of items successfully toggled
func (m *Manager) TogglePinMultiple(indices []int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(indices) == 0 {
		return 0
	}

	toggled := 0
	seen := make(map[int]bool)
	for _, idx := range indices {
		// Skip duplicates
		if seen[idx] {
			continue
		}
		seen[idx] = true

		if idx < 0 || idx >= len(m.filtered) {
			continue
		}

		histIdx, exists := m.idIndexMap[m.filtered[idx].ID]
		if !exists {
			continue
		}

		m.history[histIdx].Pinned = !m.history[histIdx].Pinned
		toggled++
	}

	if toggled > 0 {
		m.queueSaveHistory()
	}

	return toggled
}

// ValidateItem validates an item before adding
func (m *Manager) ValidateItem(item *Item) error {
	return ValidateItem(item)
}
