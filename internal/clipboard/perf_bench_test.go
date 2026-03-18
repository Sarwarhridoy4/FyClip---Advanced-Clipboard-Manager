package clipboard

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func makeBenchmarkItems(n int) []Item {
	items := make([]Item, 0, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		items = append(items, Item{
			ID:        fmt.Sprintf("%d", i),
			Type:      TypeText,
			Content:   fmt.Sprintf("Clip %d %s", i, strings.Repeat("text ", 8)),
			Timestamp: now.Add(-time.Duration(i) * time.Second),
			Pinned:    i%10 == 0,
			Hash:      fmt.Sprintf("hash-%d", i),
		})
	}
	return items
}

func BenchmarkUpdateFilteredSearch1000(b *testing.B) {
	m := &Manager{
		history:      makeBenchmarkItems(1000),
		hashIndexMap: make(map[string]int, 1000),
		idIndexMap:   make(map[string]int, 1000),
	}
	// Build index maps
	for i, item := range m.history {
		m.hashIndexMap[item.Hash+":"+strconv.Itoa(int(item.Type))] = i
		m.idIndexMap[item.ID] = i
	}
	m.searchQuery = "clip 99"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.updateFiltered()
	}
}

func BenchmarkAddItemWithDuplicateScan1000(b *testing.B) {
	m := &Manager{
		history:      makeBenchmarkItems(1000),
		hashIndexMap: make(map[string]int, 1000),
		idIndexMap:   make(map[string]int, 1000),
		maxHistoryItems: 1000,
	}
	// Build index maps
	for i, item := range m.history {
		m.hashIndexMap[item.Hash+":"+strconv.Itoa(int(item.Type))] = i
		m.idIndexMap[item.ID] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := Item{
			Type:    TypeText,
			Content: fmt.Sprintf("new clip %d", i),
			Hash:    fmt.Sprintf("new-hash-%d", i),
		}
		m.AddItem(item)
		if len(m.history) > 1000 {
			m.history = m.history[len(m.history)-1000:]
			// Rebuild index maps after truncation
			m.hashIndexMap = make(map[string]int, 1000)
			m.idIndexMap = make(map[string]int, 1000)
			for j, it := range m.history {
				m.hashIndexMap[it.Hash+":"+strconv.Itoa(int(it.Type))] = j
				m.idIndexMap[it.ID] = j
			}
		}
	}
}

func BenchmarkStorageSave1000(b *testing.B) {
	dir := b.TempDir()
	s, err := NewStorage(dir)
	if err != nil {
		b.Fatalf("new storage: %v", err)
	}
	items := makeBenchmarkItems(1000)
	_ = filepath.Join(dir, "noop")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := s.Save(items); err != nil {
			b.Fatalf("save: %v", err)
		}
	}
}
