// File: internal/metrics/metrics.go
package metrics

import (
	"expvar"
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mu sync.RWMutex

	// Clipboard metrics
	ItemsAdded      int64
	ItemsDeleted    int64
	ItemsPinned     int64
	ItemsUnpinned   int64
	Searches        int64
	SearchErrors    int64
	Saves           int64
	SaveErrors      int64
	Loads           int64
	LoadErrors      int64

	// Performance metrics
	AverageSearchDuration time.Duration
	AverageSaveDuration   time.Duration
	AverageLoadDuration   time.Duration
	TotalSearchDuration   time.Duration
	TotalSaveDuration     time.Duration
	TotalLoadDuration     time.Duration

	// History metrics
	CurrentHistorySize int64
	MaxHistorySize     int64

	// Error metrics
	Errors int64

	// Type distribution
	TextItems   int64
	ImageItems  int64
	HTMLItems   int64
	FileItems   int64
}

// Global metrics instance
var global = &Metrics{}

// ExpVar metrics exposed via /debug/vars
var (
	expItemsAdded     = expvar.NewInt("clipboard_items_added")
	expItemsDeleted   = expvar.NewInt("clipboard_items_deleted")
	expItemsPinned    = expvar.NewInt("clipboard_items_pinned")
	expItemsUnpinned = expvar.NewInt("clipboard_items_unpinned")
	expSearches       = expvar.NewInt("clipboard_searches")
	expSearchErrors   = expvar.NewInt("clipboard_search_errors")
	expSaves          = expvar.NewInt("clipboard_saves")
	expSaveErrors     = expvar.NewInt("clipboard_save_errors")
	expLoads          = expvar.NewInt("clipboard_loads")
	expLoadErrors     = expvar.NewInt("clipboard_load_errors")
	expErrors         = expvar.NewInt("clipboard_errors")
	expHistorySize    = expvar.NewInt("clipboard_history_size")
	expTextItems      = expvar.NewInt("clipboard_text_items")
	expImageItems     = expvar.NewInt("clipboard_image_items")
	expHTMLItems      = expvar.NewInt("clipboard_html_items")
	expFileItems      = expvar.NewInt("clipboard_file_items")
)

// Get returns the global metrics instance
func Get() *Metrics {
	return global
}

// RecordItemAdded records an item addition
func RecordItemAdded() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.ItemsAdded++
	expItemsAdded.Add(1)
}

// RecordItemDeleted records an item deletion
func RecordItemDeleted() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.ItemsDeleted++
	expItemsDeleted.Add(1)
}

// RecordItemPinned records a pin operation
func RecordItemPinned() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.ItemsPinned++
	expItemsPinned.Add(1)
}

// RecordItemUnpinned records an unpin operation
func RecordItemUnpinned() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.ItemsUnpinned++
	expItemsUnpinned.Add(1)
}

// RecordSearch records a search operation
func RecordSearch(duration time.Duration) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.Searches++
	global.TotalSearchDuration += duration
	if global.Searches > 0 {
		global.AverageSearchDuration = global.TotalSearchDuration / time.Duration(global.Searches)
	}
	expSearches.Add(1)
}

// RecordSearchError records a search error
func RecordSearchError() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.SearchErrors++
	expSearchErrors.Add(1)
}

// RecordSave records a save operation
func RecordSave(duration time.Duration) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.Saves++
	global.TotalSaveDuration += duration
	if global.Saves > 0 {
		global.AverageSaveDuration = global.TotalSaveDuration / time.Duration(global.Saves)
	}
	expSaves.Add(1)
}

// RecordSaveError records a save error
func RecordSaveError() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.SaveErrors++
	expSaveErrors.Add(1)
}

// RecordLoad records a load operation
func RecordLoad(duration time.Duration) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.Loads++
	global.TotalLoadDuration += duration
	if global.Loads > 0 {
		global.AverageLoadDuration = global.TotalLoadDuration / time.Duration(global.Loads)
	}
	expLoads.Add(1)
}

// RecordLoadError records a load error
func RecordLoadError() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.LoadErrors++
	expLoadErrors.Add(1)
}

// RecordError records a general error
func RecordError() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.Errors++
	expErrors.Add(1)
}

// UpdateHistorySize updates the current history size
func UpdateHistorySize(size int) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.CurrentHistorySize = int64(size)
	expHistorySize.Set(int64(size))
}

// UpdateMaxHistorySize updates the max history size
func UpdateMaxHistorySize(size int) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.MaxHistorySize = int64(size)
}

// UpdateTypeDistribution updates the type distribution counts
func UpdateTypeDistribution(text, image, html, file int) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.TextItems = int64(text)
	global.ImageItems = int64(image)
	global.HTMLItems = int64(html)
	global.FileItems = int64(file)
	expTextItems.Set(int64(text))
	expImageItems.Set(int64(image))
	expHTMLItems.Set(int64(html))
	expFileItems.Set(int64(file))
}

// GetStats returns current metrics as a map
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"items_added":           m.ItemsAdded,
		"items_deleted":         m.ItemsDeleted,
		"items_pinned":         m.ItemsPinned,
		"items_unpinned":       m.ItemsUnpinned,
		"searches":             m.Searches,
		"search_errors":        m.SearchErrors,
		"saves":                m.Saves,
		"save_errors":          m.SaveErrors,
		"loads":                m.Loads,
		"load_errors":          m.LoadErrors,
		"errors":               m.Errors,
		"history_size":         m.CurrentHistorySize,
		"max_history_size":     m.MaxHistorySize,
		"avg_search_duration":  m.AverageSearchDuration.String(),
		"avg_save_duration":    m.AverageSaveDuration.String(),
		"avg_load_duration":    m.AverageLoadDuration.String(),
		"text_items":           m.TextItems,
		"image_items":          m.ImageItems,
		"html_items":           m.HTMLItems,
		"file_items":           m.FileItems,
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ItemsAdded = 0
	m.ItemsDeleted = 0
	m.ItemsPinned = 0
	m.ItemsUnpinned = 0
	m.Searches = 0
	m.SearchErrors = 0
	m.Saves = 0
	m.SaveErrors = 0
	m.Loads = 0
	m.LoadErrors = 0
	m.Errors = 0
	m.AverageSearchDuration = 0
	m.AverageSaveDuration = 0
	m.AverageLoadDuration = 0
	m.TotalSearchDuration = 0
	m.TotalSaveDuration = 0
	m.TotalLoadDuration = 0

	// Reset expvars
	expItemsAdded.Set(0)
	expItemsDeleted.Set(0)
	expItemsPinned.Set(0)
	expItemsUnpinned.Set(0)
	expSearches.Set(0)
	expSearchErrors.Set(0)
	expSaves.Set(0)
	expSaveErrors.Set(0)
	expLoads.Set(0)
	expLoadErrors.Set(0)
	expErrors.Set(0)
	expHistorySize.Set(0)
	expTextItems.Set(0)
	expImageItems.Set(0)
	expHTMLItems.Set(0)
	expFileItems.Set(0)
}
