// File: internal/clipboard/snippet.go
package clipboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Snippet represents a text snippet/template
type Snippet struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Abbreviation string    `json:"abbreviation,omitempty"` // Short trigger like "sig"
	Category     string    `json:"category,omitempty"`
	IsSystem     bool      `json:"is_system,omitempty"` // System snippets cannot be deleted
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SnippetManager handles snippet storage and operations
type SnippetManager struct {
	snippets []Snippet
	storage  *Storage
}

// NewSnippetManager creates a new snippet manager
func NewSnippetManager(storage *Storage) *SnippetManager {
	return &SnippetManager{
		snippets: []Snippet{},
		storage:  storage,
	}
}

// LoadSnippets loads snippets from storage
func (sm *SnippetManager) LoadSnippets() error {
	// Try to load snippets - they may not exist yet
	// We'll store snippets in a separate file
	path := filepath.Join(sm.storage.GetDir(), "snippets.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No snippets file yet, start with defaults
			sm.snippets = getDefaultSnippets()
			return sm.SaveSnippets()
		}
		return err
	}
	
	return json.Unmarshal(data, &sm.snippets)
}

// SaveSnippets saves snippets to storage
func (sm *SnippetManager) SaveSnippets() error {
	path := filepath.Join(sm.storage.GetDir(), "snippets.json")
	data, err := json.MarshalIndent(sm.snippets, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// GetSnippets returns all snippets
func (sm *SnippetManager) GetSnippets() []Snippet {
	return sm.snippets
}

// GetSnippetByID returns a snippet by ID
func (sm *SnippetManager) GetSnippetByID(id string) (Snippet, bool) {
	for _, s := range sm.snippets {
		if s.ID == id {
			return s, true
		}
	}
	return Snippet{}, false
}

// GetSnippetByAbbreviation returns a snippet by abbreviation
func (sm *SnippetManager) GetSnippetByAbbreviation(abbr string) (Snippet, bool) {
	for _, s := range sm.snippets {
		if s.Abbreviation == abbr {
			return s, true
		}
	}
	return Snippet{}, false
}

// AddSnippet adds a new snippet
func (sm *SnippetManager) AddSnippet(snippet Snippet) error {
	snippet.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	snippet.CreatedAt = time.Now()
	snippet.UpdatedAt = time.Now()
	
	sm.snippets = append(sm.snippets, snippet)
	return sm.SaveSnippets()
}

// UpdateSnippet updates an existing snippet
func (sm *SnippetManager) UpdateSnippet(snippet Snippet) error {
	for i, s := range sm.snippets {
		if s.ID == snippet.ID {
			snippet.UpdatedAt = time.Now()
			snippet.CreatedAt = s.CreatedAt
			sm.snippets[i] = snippet
			return sm.SaveSnippets()
		}
	}
	return fmt.Errorf("snippet not found")
}

// DeleteSnippet deletes a snippet by ID
func (sm *SnippetManager) DeleteSnippet(id string) error {
	for i, s := range sm.snippets {
		if s.ID == id {
			sm.snippets = append(sm.snippets[:i], sm.snippets[i+1:]...)
			return sm.SaveSnippets()
		}
	}
	return fmt.Errorf("snippet not found")
}

// ExpandSnippet expands template variables in snippet content
func (sm *SnippetManager) ExpandSnippet(content string, clipboardContent string) string {
	result := content
	
	// Replace template variables
	now := time.Now()
	
	result = strings.ReplaceAll(result, "{{date}}", now.Format("2006-01-02"))
	result = strings.ReplaceAll(result, "{{time}}", now.Format("15:04:05"))
	result = strings.ReplaceAll(result, "{{datetime}}", now.Format("2006-01-02 15:04:05"))
	result = strings.ReplaceAll(result, "{{clipboard}}", clipboardContent)
	
	// Add more date/time formats
	result = strings.ReplaceAll(result, "{{year}}", fmt.Sprintf("%d", now.Year()))
	result = strings.ReplaceAll(result, "{{month}}", now.Format("01"))
	result = strings.ReplaceAll(result, "{{day}}", now.Format("02"))
	
	return result
}

// GetCategories returns all unique categories
func (sm *SnippetManager) GetCategories() []string {
	catMap := make(map[string]bool)
	for _, s := range sm.snippets {
		if s.Category != "" {
			catMap[s.Category] = true
		}
	}
	
	categories := []string{}
	for cat := range catMap {
		categories = append(categories, cat)
	}
	
	return categories
}

// getDefaultSnippets returns some default snippets (system snippets that cannot be deleted)
func getDefaultSnippets() []Snippet {
	now := time.Now()
	return []Snippet{
		{
			Title:        "Email Signature",
			Content:      "Best regards,\n{{date}}",
			Abbreviation: "sig",
			Category:     "General",
			IsSystem:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			Title:        "Current Date",
			Content:      "{{date}}",
			Abbreviation: "date",
			Category:     "Utility",
			IsSystem:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			Title:        "Current DateTime",
			Content:      "{{datetime}}",
			Abbreviation: "dt",
			Category:     "Utility",
			IsSystem:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			Title:        "Current Time",
			Content:      "{{time}}",
			Abbreviation: "time",
			Category:     "Utility",
			IsSystem:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			Title:        "Clipboard Content",
			Content:      "{{clipboard}}",
			Abbreviation: "clip",
			Category:     "Utility",
			IsSystem:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}
