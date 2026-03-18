// File: internal/clipboard/exclusion.go
package clipboard

import (
	"regexp"
	"strconv"
	"strings"
)

// ExclusionRuleType represents the type of exclusion rule
type ExclusionRuleType string

const (
	ExclusionTypeRegex ExclusionRuleType = "regex"
	ExclusionTypeApp   ExclusionRuleType = "app"
	ExclusionTypeSize  ExclusionRuleType = "size"
)

// ExclusionRule represents a rule for excluding clipboard content
type ExclusionRule struct {
	ID      string           `json:"id"`
	Type    ExclusionRuleType `json:"type"`   // "regex", "app", "size"
	Pattern string           `json:"pattern"` // regex pattern or app name
	Value   string           `json:"value"`   // size in MB for type "size"
	Enabled bool             `json:"enabled"`
}

// ExclusionManager handles exclusion rules for clipboard monitoring
type ExclusionManager struct {
	rules []ExclusionRule
}

// NewExclusionManager creates a new exclusion manager
func NewExclusionManager() *ExclusionManager {
	return &ExclusionManager{
		rules: getDefaultExclusionRules(),
	}
}

// LoadRules loads exclusion rules from config
func (em *ExclusionManager) LoadRules(rules []ExclusionRule) {
	if len(rules) == 0 {
		em.rules = getDefaultExclusionRules()
	} else {
		em.rules = rules
	}
}

// GetRules returns all exclusion rules
func (em *ExclusionManager) GetRules() []ExclusionRule {
	return em.rules
}

// AddRule adds a new exclusion rule
func (em *ExclusionManager) AddRule(rule ExclusionRule) {
	em.rules = append(em.rules, rule)
}

// RemoveRule removes an exclusion rule by ID
func (em *ExclusionManager) RemoveRule(id string) {
	for i, r := range em.rules {
		if r.ID == id {
			em.rules = append(em.rules[:i], em.rules[i+1:]...)
			return
		}
	}
}

// UpdateRule updates an existing rule
func (em *ExclusionManager) UpdateRule(rule ExclusionRule) {
	for i, r := range em.rules {
		if r.ID == rule.ID {
			em.rules[i] = rule
			return
		}
	}
}

// ShouldExclude checks if content should be excluded based on rules
func (em *ExclusionManager) ShouldExclude(content string, contentSize int, appName string) (bool, string) {
	for _, rule := range em.rules {
		if !rule.Enabled {
			continue
		}

		switch rule.Type {
		case ExclusionTypeRegex:
			if content == "" {
				continue
			}
			// Compile and check regex
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue // Invalid regex, skip
			}
			if re.MatchString(content) {
				return true, "matches exclusion pattern: " + rule.Pattern
			}

		case ExclusionTypeApp:
			if appName == "" {
				continue
			}
			// Check if app name contains the pattern (case insensitive)
			if strings.Contains(strings.ToLower(appName), strings.ToLower(rule.Pattern)) {
				return true, "from excluded app: " + rule.Pattern
			}

		case ExclusionTypeSize:
			maxSizeMB, err := strconv.Atoi(rule.Value)
			if err != nil {
				continue // Invalid size, skip
			}
			sizeMB := contentSize / (1024 * 1024)
			if sizeMB >= maxSizeMB {
				return true, "exceeds size limit: " + rule.Value + "MB"
			}
		}
	}

	return false, ""
}

// getDefaultExclusionRules returns default exclusion rules
func getDefaultExclusionRules() []ExclusionRule {
	return []ExclusionRule{
		{
			ID:      "default_password_managers",
			Type:    ExclusionTypeApp,
			Pattern: "1password",
			Value:   "",
			Enabled: true,
		},
		{
			ID:      "default_bitwarden",
			Type:    ExclusionTypeApp,
			Pattern: "bitwarden",
			Value:   "",
			Enabled: true,
		},
		{
			ID:      "default_keepass",
			Type:    ExclusionTypeApp,
			Pattern: "keepass",
			Value:   "",
			Enabled: true,
		},
		{
			ID:      "default_large_size",
			Type:    ExclusionTypeSize,
			Pattern: "",
			Value:   "10",
			Enabled: true,
		},
	}
}
