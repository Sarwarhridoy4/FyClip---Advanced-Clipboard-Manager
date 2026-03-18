// File: internal/clipboard/sensitive.go
package clipboard

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"
)

// SensitiveFlag marks items as sensitive (never save certain content)
const SensitiveFlag = "sensitive"

// Default sensitive patterns to detect
var defaultSensitivePatterns = []*regexp.Regexp{
	// Credit card patterns (simple regex)
	regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
	// SSN pattern (US)
	regexp.MustCompile(`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`),
	// API keys / tokens (common patterns)
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|auth)[=:\s]+\S+`),
	// Private keys
	regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
}

// ExclusionManager manages pattern exclusion rules
type SensitiveManager struct {
	mu            sync.RWMutex
	patterns      []*regexp.Regexp
	enabled       bool
	sensitiveIDs  map[string]bool
	passwordApps  []string
}

// NewSensitiveManager creates a new sensitive data manager
func NewSensitiveManager() *SensitiveManager {
	return &SensitiveManager{
		patterns:     defaultSensitivePatterns,
		enabled:      true,
		sensitiveIDs: make(map[string]bool),
		passwordApps: []string{
			"1password",
			"bitwarden",
			"keepass",
			"keepassxc",
			"lastpass",
			"dashlane",
			"enpass",
			"nordpass",
		},
	}
}

// IsSensitive checks if content contains sensitive data
func (sm *SensitiveManager) IsSensitive(content string) bool {
	if !sm.enabled {
		return false
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, pattern := range sm.patterns {
		if pattern.MatchString(content) {
			return true
		}
	}

	return false
}

// MarkSensitive marks an item as sensitive by ID
func (sm *SensitiveManager) MarkSensitive(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sensitiveIDs[id] = true
}

// IsMarkedSensitive checks if an item is marked as sensitive
func (sm *SensitiveManager) IsMarkedSensitive(id string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sensitiveIDs[id]
}

// UnmarkSensitive removes sensitive marking from an item
func (sm *SensitiveManager) UnmarkSensitive(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sensitiveIDs, id)
}

// AddPattern adds a new sensitive pattern
func (sm *SensitiveManager) AddPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.patterns = append(sm.patterns, re)
	return nil
}

// RemovePattern removes a sensitive pattern
func (sm *SensitiveManager) RemovePattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, p := range sm.patterns {
		if p.String() == re.String() {
			sm.patterns = append(sm.patterns[:i], sm.patterns[i+1:]...)
			return nil
		}
	}

	return nil
}

// SetEnabled enables or disables sensitive data detection
func (sm *SensitiveManager) SetEnabled(enabled bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.enabled = enabled
}

// IsEnabled returns if sensitive data detection is enabled
func (sm *SensitiveManager) IsEnabled() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.enabled
}

// GetPatterns returns all registered sensitive patterns
func (sm *SensitiveManager) GetPatterns() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	patterns := make([]string, len(sm.patterns))
	for i, p := range sm.patterns {
		patterns[i] = p.String()
	}
	return patterns
}

// SecureWipe securely wipes sensitive data from an item
func (i *Item) SecureWipe() {
	if i == nil {
		return
	}

	// Clear content
	if len(i.Content) > 0 {
		i.Content = strings.Repeat("\x00", len(i.Content))
	}

	// Clear HTML content
	if len(i.HTMLContent) > 0 {
		i.HTMLContent = strings.Repeat("\x00", len(i.HTMLContent))
	}

	// Clear image data
	if len(i.ImageData) > 0 {
		i.ImageData = strings.Repeat("\x00", len(i.ImageData))
	}

	// Clear file info
	if i.FileInfo != nil {
		i.FileInfo.Path = strings.Repeat("\x00", len(i.FileInfo.Path))
		i.FileInfo.Name = strings.Repeat("\x00", len(i.FileInfo.Name))
	}

	// Clear hash
	i.Hash = ""
}

// GenerateSecureID generates a cryptographically secure random ID
func GenerateSecureID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to random hex
		return hex.EncodeToString(bytes)
	}
	return hex.EncodeToString(bytes)
}
