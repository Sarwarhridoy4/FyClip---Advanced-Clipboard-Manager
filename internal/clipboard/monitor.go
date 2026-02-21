// File: internal/clipboard/monitor.go
package clipboard

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

const monitorInterval = 300 * time.Millisecond

// Monitor watches the clipboard for changes
type Monitor struct {
	manager *Manager
	native  *NativeClipboard

	mu            sync.RWMutex
	lastTextHash  string
	lastImageHash string

	programmaticCopy bool
	programmaticHash string
	pausedUntil      time.Time

	stopChan chan struct{}
	running  bool
}

// NewMonitor creates a new clipboard monitor
func NewMonitor(manager *Manager, native *NativeClipboard) *Monitor {
	return &Monitor{
		manager:  manager,
		native:   native,
		stopChan: make(chan struct{}),
	}
}

// Start begins monitoring the clipboard
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop()
}

// Stop stops monitoring the clipboard
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	select {
	case <-m.stopChan:
	default:
		close(m.stopChan)
	}
}

// IsRunning returns whether the monitor is running
func (m *Monitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// PauseFor pauses clipboard capture for the provided duration.
func (m *Monitor) PauseFor(d time.Duration) {
	if d <= 0 {
		return
	}
	m.mu.Lock()
	m.pausedUntil = time.Now().Add(d)
	m.mu.Unlock()
}

// Resume clears any active monitor pause.
func (m *Monitor) Resume() {
	m.mu.Lock()
	m.pausedUntil = time.Time{}
	m.mu.Unlock()
}

// IsPaused reports whether monitoring is currently paused.
func (m *Monitor) IsPaused() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return !m.pausedUntil.IsZero() && time.Now().Before(m.pausedUntil)
}

// SetProgrammaticCopy marks the next clipboard change as programmatic
func (m *Monitor) SetProgrammaticCopy(data []byte) {
	hash := sha256.Sum256(data)

	m.mu.Lock()
	m.programmaticCopy = true
	m.programmaticHash = hex.EncodeToString(hash[:])
	m.mu.Unlock()

	// Clear after a short delay
	time.AfterFunc(200*time.Millisecond, func() {
		m.mu.Lock()
		m.programmaticCopy = false
		m.programmaticHash = ""
		m.mu.Unlock()
	})
}

// monitorLoop continuously checks the clipboard
func (m *Monitor) monitorLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in monitor loop: %v", r)
			// Attempt restart after panic
			time.Sleep(1 * time.Second)
			m.mu.RLock()
			running := m.running
			m.mu.RUnlock()
			if running {
				go m.monitorLoop()
			}
		}
	}()

	ticker := time.NewTicker(monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkClipboard()

		case <-m.stopChan:
			return
		}
	}
}

// checkClipboard checks for clipboard changes
func (m *Monitor) checkClipboard() {
	m.mu.RLock()
	isProgrammatic := m.programmaticCopy
	programmaticHash := m.programmaticHash
	pausedUntil := m.pausedUntil
	m.mu.RUnlock()

	if isProgrammatic {
		return
	}

	if !pausedUntil.IsZero() {
		if time.Now().Before(pausedUntil) {
			return
		}
		m.mu.Lock()
		m.pausedUntil = time.Time{}
		m.mu.Unlock()
	}

	// Try reading text first
	if textData := m.native.ReadText(); len(textData) > 0 {
		m.handleText(textData, programmaticHash)
		return
	}

	// Try reading image
	if imageData, imageType := m.native.ReadImage(); len(imageData) > 0 {
		m.handleImage(imageData, imageType, programmaticHash)
	}
}

// handleText processes text clipboard content
func (m *Monitor) handleText(data []byte, programmaticHash string) {
	content := string(data)
	if len(content) == 0 {
		return
	}

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	m.mu.Lock()
	lastHash := m.lastTextHash
	m.mu.Unlock()

	if hashStr == lastHash || hashStr == programmaticHash {
		return
	}

	m.mu.Lock()
	m.lastTextHash = hashStr
	m.lastImageHash = "" // Clear image hash when text is copied
	m.mu.Unlock()

	item := Item{
		Type:    TypeText,
		Content: content,
		Hash:    hashStr,
	}

	result := m.manager.AddItem(item)
	if result.Added {
		if result.MovedDuplicate {
			m.manager.notifyInfo("Duplicate item moved to latest")
		}
		m.manager.saveHistory()
		m.manager.updateFiltered()
		m.manager.triggerUpdate()
	}
}

// handleImage processes image clipboard content
func (m *Monitor) handleImage(data []byte, imageType, programmaticHash string) {
	if len(data) == 0 {
		return
	}

	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	m.mu.Lock()
	lastHash := m.lastImageHash
	m.mu.Unlock()

	if hashStr == lastHash || hashStr == programmaticHash {
		return
	}

	m.mu.Lock()
	m.lastImageHash = hashStr
	m.lastTextHash = "" // Clear text hash when image is copied
	m.mu.Unlock()

	item := Item{
		Type:      TypeImage,
		Content:   fmt.Sprintf("Image (%d bytes)", len(data)),
		ImageData: base64.StdEncoding.EncodeToString(data),
		ImageType: imageType,
		Hash:      hashStr,
	}

	result := m.manager.AddItem(item)
	if result.Added {
		if result.MovedDuplicate {
			m.manager.notifyInfo("Duplicate item moved to latest")
		}
		m.manager.saveHistory()
		m.manager.updateFiltered()
		m.manager.triggerUpdate()
	}
}
