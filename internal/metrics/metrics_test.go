// File: internal/metrics/metrics_test.go
package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	// Get should return the same instance
	m1 := Get()
	m2 := Get()
	
	if m1 != m2 {
		t.Error("Get() should return the same metrics instance")
	}
}

func TestRecordItemAdded(t *testing.T) {
	m := &Metrics{}
	global = m // Reset global for testing
	
	initial := m.ItemsAdded
	RecordItemAdded()
	
	if m.ItemsAdded != initial+1 {
		t.Errorf("ItemsAdded = %d, want %d", m.ItemsAdded, initial+1)
	}
}

func TestRecordItemDeleted(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.ItemsDeleted
	RecordItemDeleted()
	
	if m.ItemsDeleted != initial+1 {
		t.Errorf("ItemsDeleted = %d, want %d", m.ItemsDeleted, initial+1)
	}
}

func TestRecordItemPinned(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.ItemsPinned
	RecordItemPinned()
	
	if m.ItemsPinned != initial+1 {
		t.Errorf("ItemsPinned = %d, want %d", m.ItemsPinned, initial+1)
	}
}

func TestRecordItemUnpinned(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.ItemsUnpinned
	RecordItemUnpinned()
	
	if m.ItemsUnpinned != initial+1 {
		t.Errorf("ItemsUnpinned = %d, want %d", m.ItemsUnpinned, initial+1)
	}
}

func TestRecordSearch(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.Searches
	duration := 100 * time.Millisecond
	RecordSearch(duration)
	
	if m.Searches != initial+1 {
		t.Errorf("Searches = %d, want %d", m.Searches, initial+1)
	}
	
	if m.TotalSearchDuration != duration {
		t.Errorf("TotalSearchDuration = %v, want %v", m.TotalSearchDuration, duration)
	}
	
	// Second search should update average
	duration2 := 200 * time.Millisecond
	RecordSearch(duration2)
	
	// Average should be (100ms + 200ms) / 2 = 150ms
	expectedAvg := (duration + duration2) / 2
	if m.AverageSearchDuration != expectedAvg {
		t.Errorf("AverageSearchDuration = %v, want %v", m.AverageSearchDuration, expectedAvg)
	}
}

func TestRecordSearchError(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.SearchErrors
	RecordSearchError()
	
	if m.SearchErrors != initial+1 {
		t.Errorf("SearchErrors = %d, want %d", m.SearchErrors, initial+1)
	}
}

func TestRecordSave(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.Saves
	duration := 50 * time.Millisecond
	RecordSave(duration)
	
	if m.Saves != initial+1 {
		t.Errorf("Saves = %d, want %d", m.Saves, initial+1)
	}
	
	if m.TotalSaveDuration != duration {
		t.Errorf("TotalSaveDuration = %v, want %v", m.TotalSaveDuration, duration)
	}
}

func TestRecordSaveError(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.SaveErrors
	RecordSaveError()
	
	if m.SaveErrors != initial+1 {
		t.Errorf("SaveErrors = %d, want %d", m.SaveErrors, initial+1)
	}
}

func TestRecordLoad(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.Loads
	duration := 75 * time.Millisecond
	RecordLoad(duration)
	
	if m.Loads != initial+1 {
		t.Errorf("Loads = %d, want %d", m.Loads, initial+1)
	}
	
	if m.TotalLoadDuration != duration {
		t.Errorf("TotalLoadDuration = %v, want %v", m.TotalLoadDuration, duration)
	}
}

func TestRecordLoadError(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.LoadErrors
	RecordLoadError()
	
	if m.LoadErrors != initial+1 {
		t.Errorf("LoadErrors = %d, want %d", m.LoadErrors, initial+1)
	}
}

func TestRecordError(t *testing.T) {
	m := &Metrics{}
	global = m
	
	initial := m.Errors
	RecordError()
	
	if m.Errors != initial+1 {
		t.Errorf("Errors = %d, want %d", m.Errors, initial+1)
	}
}

func TestUpdateHistorySize(t *testing.T) {
	m := &Metrics{}
	global = m
	
	UpdateHistorySize(50)
	
	if m.CurrentHistorySize != 50 {
		t.Errorf("CurrentHistorySize = %d, want 50", m.CurrentHistorySize)
	}
	
	UpdateHistorySize(100)
	
	if m.CurrentHistorySize != 100 {
		t.Errorf("CurrentHistorySize = %d, want 100", m.CurrentHistorySize)
	}
}

func TestUpdateMaxHistorySize(t *testing.T) {
	m := &Metrics{}
	global = m
	
	UpdateMaxHistorySize(1000)
	
	if m.MaxHistorySize != 1000 {
		t.Errorf("MaxHistorySize = %d, want 1000", m.MaxHistorySize)
	}
}

func TestUpdateTypeDistribution(t *testing.T) {
	m := &Metrics{}
	global = m
	
	UpdateTypeDistribution(10, 5, 3, 2)
	
	if m.TextItems != 10 {
		t.Errorf("TextItems = %d, want 10", m.TextItems)
	}
	if m.ImageItems != 5 {
		t.Errorf("ImageItems = %d, want 5", m.ImageItems)
	}
	if m.HTMLItems != 3 {
		t.Errorf("HTMLItems = %d, want 3", m.HTMLItems)
	}
	if m.FileItems != 2 {
		t.Errorf("FileItems = %d, want 2", m.FileItems)
	}
}

func TestGetStats(t *testing.T) {
	m := &Metrics{
		ItemsAdded:   10,
		ItemsDeleted: 5,
		Searches:    20,
		Errors:      2,
	}
	
	stats := m.GetStats()
	
	if stats["items_added"] != int64(10) {
		t.Errorf("items_added = %v, want 10", stats["items_added"])
	}
	if stats["items_deleted"] != int64(5) {
		t.Errorf("items_deleted = %v, want 5", stats["items_deleted"])
	}
	if stats["searches"] != int64(20) {
		t.Errorf("searches = %v, want 20", stats["searches"])
	}
	if stats["errors"] != int64(2) {
		t.Errorf("errors = %v, want 2", stats["errors"])
	}
}

func TestReset(t *testing.T) {
	m := &Metrics{
		ItemsAdded:          10,
		ItemsDeleted:        5,
		ItemsPinned:         3,
		ItemsUnpinned:       2,
		Searches:           20,
		SearchErrors:        1,
		Saves:               15,
		SaveErrors:          1,
		Loads:               25,
		LoadErrors:          1,
		Errors:              3,
		CurrentHistorySize:  50,
		MaxHistorySize:      1000,
		AverageSearchDuration: 100 * time.Millisecond,
		TotalSearchDuration:   2 * time.Second,
	}
	
	m.Reset()
	
	// Note: Reset() doesn't reset CurrentHistorySize or MaxHistorySize per the implementation
	// It only resets counters and durations
	if m.ItemsAdded != 0 {
		t.Errorf("ItemsAdded = %d, want 0", m.ItemsAdded)
	}
	if m.ItemsDeleted != 0 {
		t.Errorf("ItemsDeleted = %d, want 0", m.ItemsDeleted)
	}
	if m.Searches != 0 {
		t.Errorf("Searches = %d, want 0", m.Searches)
	}
	if m.Errors != 0 {
		t.Errorf("Errors = %d, want 0", m.Errors)
	}
	// These are intentionally not reset per the implementation
	_ = m.CurrentHistorySize
	_ = m.MaxHistorySize
}

func TestMetricsConcurrency(t *testing.T) {
	m := &Metrics{}
	global = m
	
	var wg sync.WaitGroup
	numGoroutines := 100
	iterations := 100
	
	// Run concurrent record operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				RecordItemAdded()
				RecordSearch(time.Millisecond)
				RecordSave(time.Millisecond)
				RecordLoad(time.Millisecond)
			}
		}()
	}
	
	wg.Wait()
	
	// Check counts
	expected := int64(numGoroutines * iterations)
	if m.ItemsAdded != expected {
		t.Errorf("ItemsAdded = %d, want %d", m.ItemsAdded, expected)
	}
	if m.Searches != expected {
		t.Errorf("Searches = %d, want %d", m.Searches, expected)
	}
	if m.Saves != expected {
		t.Errorf("Saves = %d, want %d", m.Saves, expected)
	}
	if m.Loads != expected {
		t.Errorf("Loads = %d, want %d", m.Loads, expected)
	}
}
