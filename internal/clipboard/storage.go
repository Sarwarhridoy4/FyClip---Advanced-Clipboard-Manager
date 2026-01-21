// File: internal/clipboard/storage.go
package clipboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const historyFileName = "clipboard_history.json"

// Storage handles persistence of clipboard history
type Storage struct {
	mu       sync.RWMutex
	filePath string
}

// NewStorage creates a new storage instance
func NewStorage(basePath string) (*Storage, error) {
	if basePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		basePath = homeDir
	}

	// Create directory if it doesn't exist
	configDir := filepath.Join(basePath, ".fyclip")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &Storage{
		filePath: filepath.Join(configDir, historyFileName),
	}, nil
}

// Load reads clipboard history from disk
func (s *Storage) Load() ([]Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return items, nil
}

// Save writes clipboard history to disk
func (s *Storage) Save(items []Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write to temp file first
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile) // Cleanup
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetPath returns the storage file path
func (s *Storage) GetPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filePath
}

// Clear removes all stored history
func (s *Storage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove history file: %w", err)
	}

	return nil
}