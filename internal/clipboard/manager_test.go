// File: internal/clipboard/manager_test.go
package clipboard

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestManagerNew tests creating a new manager
func TestManagerNew(t *testing.T) {
	// Create temp directory for storage
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	
	// Verify initial state
	if m.GetFilteredCount() != 0 {
		t.Errorf("Expected 0 filtered items, got %d", m.GetFilteredCount())
	}
	
	// Cleanup
	m.Shutdown()
}

// TestManagerSearch tests search functionality
func TestManagerSearch(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Add items using updateFiltered after each add (simulating monitor behavior)
	m.AddItem(Item{Type: TypeText, Content: "Apple"})
	m.updateFiltered()
	m.AddItem(Item{Type: TypeText, Content: "Banana"})
	m.updateFiltered()
	m.AddItem(Item{Type: TypeText, Content: "Apricot"})
	m.updateFiltered()
	m.AddItem(Item{Type: TypeText, Content: "Cherry"})
	m.updateFiltered()
	
	// Search for "ap"
	m.SetSearch("ap")
	
	// Should match Apple and Apricot
	count := m.GetFilteredCount()
	if count != 2 {
		t.Errorf("Expected 2 search results, got %d", count)
	}
	
	items := m.GetFiltered()
	found := false
	for _, item := range items {
		if item.Content == "Apple" || item.Content == "Apricot" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Search results should contain Apple or Apricot")
	}
	
	// Clear search
	m.SetSearch("")
	if m.GetFilteredCount() != 4 {
		t.Errorf("Expected 4 items after clearing search, got %d", m.GetFilteredCount())
	}
}

// TestManagerSearchEmpty tests search with no matches
func TestManagerSearchEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Add items
	m.AddItem(Item{Type: TypeText, Content: "Apple"})
	m.updateFiltered()
	m.AddItem(Item{Type: TypeText, Content: "Banana"})
	m.updateFiltered()
	
	// Search for non-existent item
	m.SetSearch("xyz")
	
	if m.GetFilteredCount() != 0 {
		t.Errorf("Expected 0 search results, got %d", m.GetFilteredCount())
	}
}

// TestManagerCaseInsensitiveSearch tests case-insensitive search
func TestManagerCaseInsensitiveSearch(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Add items with different cases
	m.AddItem(Item{Type: TypeText, Content: "HELLO"})
	m.updateFiltered()
	m.AddItem(Item{Type: TypeText, Content: "hello"})
	m.updateFiltered()
	m.AddItem(Item{Type: TypeText, Content: "Hello"})
	m.updateFiltered()
	
	// Search lowercase
	m.SetSearch("hello")
	
	count := m.GetFilteredCount()
	if count != 3 {
		t.Errorf("Expected 3 results for case-insensitive search, got %d", count)
	}
}

// TestManagerPauseResume tests pause/resume monitoring
func TestManagerPauseResume(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Pause for 1 hour
	m.PauseMonitoringFor(time.Hour)
	
	if !m.IsMonitoringPaused() {
		t.Error("Expected monitoring to be paused")
	}
	
	// Resume
	m.ResumeMonitoring()
	
	if m.IsMonitoringPaused() {
		t.Error("Expected monitoring to be resumed")
	}
}

// TestManagerEmptyContent tests that empty content is rejected
func TestManagerEmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Try to add empty item
	result := m.AddItem(Item{Type: TypeText, Content: ""})
	
	if result.Added {
		t.Error("Empty content should not be added")
	}
	
	if m.GetFilteredCount() != 0 {
		t.Error("No items should be added")
	}
}

// TestManagerInvalidIndex tests handling of invalid indices
func TestManagerInvalidIndex(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Try to get item at invalid index
	_, ok := m.GetItem(-1)
	if ok {
		t.Error("Should not get item at negative index")
	}
	
	_, ok = m.GetItem(100)
	if ok {
		t.Error("Should not get item beyond bounds")
	}
	
	// Try to delete at invalid index
	err = m.Delete(-1)
	if err == nil {
		t.Error("Should error on invalid delete index")
	}
	
	// Try to toggle pin at invalid index
	success := m.TogglePin(100)
	if success {
		t.Error("Should fail toggle pin at invalid index")
	}
}

// TestManagerSetMaxHistory tests setting max history limit
func TestManagerSetMaxHistory(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Test invalid limits
	if m.SetMaxHistory(0) {
		t.Error("SetMaxHistory(0) should return false")
	}
	
	if m.SetMaxHistory(-1) {
		t.Error("SetMaxHistory(-1) should return false")
	}
	
	// Test valid limit
	if !m.SetMaxHistory(100) {
		t.Error("SetMaxHistory(100) should return true")
	}
	
	if m.GetMaxHistory() != 100 {
		t.Errorf("Expected max history 100, got %d", m.GetMaxHistory())
	}
}

// TestManagerTogglePinnedOnly tests pinned-only filter toggle
func TestManagerTogglePinnedOnly(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Initially should not be pinned-only
	if m.IsPinnedOnly() {
		t.Error("Initially should not be pinned-only")
	}
	
	// Toggle on
	enabled := m.TogglePinnedOnly()
	if !enabled {
		t.Error("TogglePinnedOnly should return true when enabling")
	}
	
	if !m.IsPinnedOnly() {
		t.Error("IsPinnedOnly should return true after toggle")
	}
	
	// Toggle off
	enabled = m.TogglePinnedOnly()
	if enabled {
		t.Error("TogglePinnedOnly should return false when disabling")
	}
	
	if m.IsPinnedOnly() {
		t.Error("IsPinnedOnly should return false after toggle")
	}
}

// TestManagerSetSearch tests search query updates
func TestManagerSetSearch(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Set search query
	m.SetSearch("test")
	
	// Verify search is set (internally)
	// The actual filtering happens in updateFiltered
	m.updateFiltered()
	
	// Clear search
	m.SetSearch("")
	m.updateFiltered()
}

// TestManagerGetSelected tests selected item retrieval
func TestManagerGetSelected(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Initially no selection
	_, ok := m.GetSelected()
	if ok {
		t.Error("Should not have selected item initially")
	}
	
	// Add item and select it
	m.AddItem(Item{Type: TypeText, Content: "Test"})
	m.updateFiltered()
	m.SetSelected(0)
	
	item, ok := m.GetSelected()
	if !ok {
		t.Error("Should have selected item")
	}
	
	if item.Content != "Test" {
		t.Errorf("Selected item content = %v; want Test", item.Content)
	}
	
	// Test GetSelectedIndex
	if m.GetSelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d; want 0", m.GetSelectedIndex())
	}
}

// TestManagerConcurrentSearchAndFilter tests concurrent search and filter operations
func TestManagerConcurrentSearchAndFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	var wg sync.WaitGroup
	
	// Concurrent search operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			query := ""
			if n%2 == 0 {
				query = "test"
			}
			m.SetSearch(query)
			m.updateFiltered()
		}(i)
	}
	
	// Concurrent read operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.GetFiltered()
			m.GetFilteredCount()
			m.GetStats()
		}()
	}
	
	wg.Wait()
	// Test passes if no race conditions or deadlocks
}


// TestManagerImageWithEmptyContent tests image with empty text content
func TestManagerImageWithEmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "test_storage")
	
	cfg := Config{
		StoragePath: storagePath,
		OnUpdate:    func() {},
		OnError:     func(err error) {},
		OnInfo:      func(message string) {},
	}
	
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer m.Shutdown()
	
	// Add image item with empty content (valid)
	item := Item{
		Type:      TypeImage,
		Content:   "", // Empty content is OK for images
		ImageData: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		ImageType: "png",
	}
	
	m.AddItem(item)
	// This should NOT be added since Content is empty and ImageData is valid
	// Looking at the code - it checks Content == "" && ImageData == "" 
	// So it should be added
	
	m.updateFiltered()
	
	if m.GetFilteredCount() != 1 {
		t.Errorf("Expected 1 image item, got %d", m.GetFilteredCount())
	}
}
