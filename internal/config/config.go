// File: internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds all application settings
type Config struct {
	MaxHistoryItems     int           `json:"max_history_items"`
	MonitoringInterval  time.Duration `json:"monitoring_interval"`
	StartMinimized      bool          `json:"start_minimized"`
	Theme               string        `json:"theme"` // "dark", "light", "system"
	PauseDuration       time.Duration `json:"pause_duration"`
	AutoSaveInterval    time.Duration `json:"auto_save_interval"`
	EnableNotifications bool          `json:"enable_notifications"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxHistoryItems:     1000,
		MonitoringInterval:  500 * time.Millisecond,
		StartMinimized:      false,
		Theme:               "dark",
		PauseDuration:       30 * time.Second,
		AutoSaveInterval:    5 * time.Second,
		EnableNotifications: true,
	}
}

// ConfigManager handles loading and saving configuration
type ConfigManager struct {
	configPath string
	config     *Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	cm := &ConfigManager{}
	
	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".fyclip")
	
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	cm.configPath = filepath.Join(configDir, "config.json")
	
	// Load or create config
	cm.config = DefaultConfig()
	if err := cm.Load(); err != nil {
		// If loading fails, save the default config
		if saveErr := cm.Save(); saveErr != nil {
			return nil, fmt.Errorf("failed to save default config: %w", saveErr)
		}
	}
	
	return cm, nil
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, use defaults
			cm.config = DefaultConfig()
			return cm.Save()
		}
		return fmt.Errorf("failed to read config: %w", err)
	}
	
	// Parse JSON with defaults for missing fields
	if err := json.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	// Apply migration if needed
	cm.migrate()
	
	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *Config {
	return cm.config
}

// Update updates configuration and saves to file
func (cm *ConfigManager) Update(cfg *Config) error {
	cm.config = cfg
	return cm.Save()
}

// migrate applies any necessary configuration migrations
func (cm *ConfigManager) migrate() {
	// Set defaults for any new fields that might not exist in older configs
	if cm.config.MaxHistoryItems == 0 {
		cm.config.MaxHistoryItems = 1000
	}
	if cm.config.MonitoringInterval == 0 {
		cm.config.MonitoringInterval = 500 * time.Millisecond
	}
	if cm.config.PauseDuration == 0 {
		cm.config.PauseDuration = 30 * time.Second
	}
	if cm.config.AutoSaveInterval == 0 {
		cm.config.AutoSaveInterval = 5 * time.Second
	}
	if cm.config.Theme == "" {
		cm.config.Theme = "dark"
	}
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}
