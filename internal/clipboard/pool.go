// File: internal/clipboard/pool.go
package clipboard

import (
	"sync"
	"time"
)

// ItemPool reuses Item objects to reduce allocations
type ItemPool struct {
	pool sync.Pool
}

// NewItemPool creates a new item pool
func NewItemPool() *ItemPool {
	return &ItemPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Item{}
			},
		},
	}
}

// Get retrieves an Item from the pool
func (p *ItemPool) Get() *Item {
	return p.pool.Get().(*Item)
}

// Put returns an Item to the pool
func (p *ItemPool) Put(item *Item) {
	if item == nil {
		return
	}

	// Reset all fields to prevent data leaks
	item.ID = ""
	item.Type = TypeText
	item.Content = ""
	item.ImageData = ""
	item.ImageType = ""
	item.HTMLContent = ""
	item.FileInfo = nil
	item.Timestamp = time.Time{}
	item.Pinned = false
	item.CopyCount = 0
	item.Hash = ""
	item.searchContent = ""

	p.pool.Put(item)
}

// GlobalItemPool is the package-level item pool
var globalItemPool = NewItemPool()

// GetFromPool retrieves an Item from the global pool
func GetFromPool() *Item {
	return globalItemPool.Get()
}

// ReturnToPool returns an Item to the global pool
func ReturnToPool(item *Item) {
	globalItemPool.Put(item)
}
