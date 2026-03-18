// File: internal/clipboard/exclusion_test.go
package clipboard

import (
	"testing"
)

// TestExclusionManagerNew tests creating a new exclusion manager
func TestExclusionManagerNew(t *testing.T) {
	em := NewExclusionManager()
	if em == nil {
		t.Fatal("NewExclusionManager returned nil")
	}
	
	// Should have default rules
	rules := em.GetRules()
	if len(rules) == 0 {
		t.Error("Expected default exclusion rules")
	}
}

// TestExclusionManagerLoadRules tests loading exclusion rules
func TestExclusionManagerLoadRules(t *testing.T) {
	em := NewExclusionManager()
	
	// Load custom rules
	rules := []ExclusionRule{
		{ID: "test1", Type: ExclusionTypeRegex, Pattern: "test", Enabled: true},
	}
	em.LoadRules(rules)
	
	loaded := em.GetRules()
	if len(loaded) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(loaded))
	}
}

// TestExclusionManagerLoadRulesEmpty tests loading with empty rules
func TestExclusionManagerLoadRulesEmpty(t *testing.T) {
	em := NewExclusionManager()
	
	// Load empty rules - should get defaults
	em.LoadRules([]ExclusionRule{})
	
	rules := em.GetRules()
	if len(rules) == 0 {
		t.Error("Expected default rules when loading empty")
	}
}

// TestExclusionManagerAddRule tests adding a new rule
func TestExclusionManagerAddRule(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear existing rules first
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "test-rule",
		Pattern: "test-pattern",
		Type:    ExclusionTypeRegex,
		Value:   ".*test.*",
		Enabled: true,
	}
	
	em.AddRule(rule)
	
	rules := em.GetRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	
	if rules[0].ID != "test-rule" {
		t.Errorf("Rule ID = %v; want test-rule", rules[0].ID)
	}
}

// TestExclusionManagerRemoveRule tests removing a rule
func TestExclusionManagerRemoveRule(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear and add test rule
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "remove-me",
		Pattern: "test",
		Type:    ExclusionTypeRegex,
		Value:   ".*",
		Enabled: true,
	}
	
	em.AddRule(rule)
	
	// Remove it
	em.RemoveRule("remove-me")
	
	rules := em.GetRules()
	if len(rules) != 0 {
		t.Error("Rule should have been removed")
	}
}

// TestExclusionManagerUpdateRule tests updating a rule
func TestExclusionManagerUpdateRule(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear and add test rule
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "update-test",
		Pattern: "test",
		Type:    ExclusionTypeRegex,
		Value:   "original",
		Enabled: true,
	}
	
	em.AddRule(rule)
	
	// Update it
	updatedRule := ExclusionRule{
		ID:      "update-test",
		Pattern: "test",
		Type:    ExclusionTypeRegex,
		Value:   "updated",
		Enabled: false,
	}
	
	em.UpdateRule(updatedRule)
	
	rules := em.GetRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	
	if rules[0].Value != "updated" {
		t.Errorf("Rule value = %v; want updated", rules[0].Value)
	}
	
	if rules[0].Enabled {
		t.Error("Rule should be disabled")
	}
}

// TestExclusionManagerShouldExcludeRegex tests regex exclusion
func TestExclusionManagerShouldExcludeRegex(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear and add regex rule
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "password-regex",
		Pattern: "(?i)password",
		Type:    ExclusionTypeRegex,
		Value:   "",
		Enabled: true,
	}
	
	em.AddRule(rule)
	
	// Test content with "password"
	excluded, reason := em.ShouldExclude("Enter your password here", 0, "")
	if !excluded {
		t.Error("Content with 'password' should be excluded")
	}
	if reason == "" {
		t.Error("Should have a reason for exclusion")
	}
	
	// Test content without the pattern
	excluded, _ = em.ShouldExclude("Hello world", 0, "")
	if excluded {
		t.Error("Content without pattern should not be excluded")
	}
}

// TestExclusionManagerShouldExcludeApp tests app exclusion
func TestExclusionManagerShouldExcludeApp(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear and add app rule
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "1password-app",
		Pattern: "1password",
		Type:    ExclusionTypeApp,
		Value:   "",
		Enabled: true,
	}
	
	em.AddRule(rule)
	
	// Test content from excluded app
	excluded, reason := em.ShouldExclude("secret data", 0, "1Password")
	if !excluded {
		t.Error("Content from 1Password should be excluded")
	}
	if reason == "" {
		t.Error("Should have a reason for exclusion")
	}
	
	// Test content from allowed app
	excluded, _ = em.ShouldExclude("hello", 0, "Chrome")
	if excluded {
		t.Error("Content from Chrome should not be excluded")
	}
}

// TestExclusionManagerShouldExcludeSize tests size exclusion
func TestExclusionManagerShouldExcludeSize(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear and add size rule
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "size-limit",
		Pattern: "",
		Type:    ExclusionTypeSize,
		Value:   "1", // 1 MB
		Enabled: true,
	}
	
	em.AddRule(rule)
	
	// Test large content (2 MB = 2*1024*1024 bytes)
	largeSize := 2 * 1024 * 1024
	excluded, reason := em.ShouldExclude("x", largeSize, "")
	if !excluded {
		t.Error("Large content should be excluded")
	}
	if reason == "" {
		t.Error("Should have a reason for exclusion")
	}
	
	// Test small content
	smallSize := 500 * 1024 // 500 KB
	excluded, _ = em.ShouldExclude("hello", smallSize, "")
	if excluded {
		t.Error("Small content should not be excluded by size")
	}
}

// TestExclusionManagerDisabledRule tests that disabled rules don't exclude
func TestExclusionManagerDisabledRule(t *testing.T) {
	em := NewExclusionManager()
	
	// Clear and add disabled rule
	em.LoadRules([]ExclusionRule{})
	
	rule := ExclusionRule{
		ID:      "disabled-rule",
		Pattern: "secret",
		Type:    ExclusionTypeRegex,
		Value:   "(?i)secret",
		Enabled: false, // Disabled
	}
	
	em.AddRule(rule)
	
	// Disabled rule should not exclude
	excluded, _ := em.ShouldExclude("This is a secret", 0, "")
	if excluded {
		t.Error("Disabled rule should not exclude")
	}
}

// TestExclusionManagerDefaultRules tests default exclusion rules
func TestExclusionManagerDefaultRules(t *testing.T) {
	em := NewExclusionManager()
	
	rules := em.GetRules()
	
	// Should have default rules for password managers and size
	foundPasswordManager := false
	foundSize := false
	
	for _, rule := range rules {
		if rule.Type == ExclusionTypeApp {
			foundPasswordManager = true
		}
		if rule.Type == ExclusionTypeSize {
			foundSize = true
		}
	}
	
	if !foundPasswordManager {
		t.Error("Expected default password manager rules")
	}
	
	if !foundSize {
		t.Error("Expected default size rule")
	}
}
