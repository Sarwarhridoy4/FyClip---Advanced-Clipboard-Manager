// File: internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.MaxHistoryItems != 1000 {
		t.Errorf("expected MaxHistoryItems to be 1000, got %d", cfg.MaxHistoryItems)
	}
	
	if cfg.MonitoringInterval != 500*time.Millisecond {
		t.Errorf("expected MonitoringInterval to be 500ms, got %v", cfg.MonitoringInterval)
	}
	
	if cfg.StartMinimized != false {
		t.Errorf("expected StartMinimized to be false, got %v", cfg.StartMinimized)
	}
	
	if cfg.Theme != "dark" {
		t.Errorf("expected Theme to be 'dark', got %s", cfg.Theme)
	}
	
	if cfg.PauseDuration != 30*time.Second {
		t.Errorf("expected PauseDuration to be 30s, got %v", cfg.PauseDuration)
	}
	
	if cfg.AutoSaveInterval != 5*time.Second {
		t.Errorf("expected AutoSaveInterval to be 5s, got %v", cfg.AutoSaveInterval)
	}
	
	if cfg.EnableNotifications != true {
		t.Errorf("expected EnableNotifications to be true, got %v", cfg.EnableNotifications)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: false,
		},
		{
			name: "invalid max history - too low",
			cfg: &Config{
				MaxHistoryItems:     0,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid max history - too high",
			cfg: &Config{
				MaxHistoryItems:     100001,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid monitoring interval - too low",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  50 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid monitoring interval - too high",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  15 * time.Second,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid pause duration - too low",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       500 * time.Millisecond,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid pause duration - too high",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       25 * time.Hour,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid auto save interval - too low",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    500 * time.Millisecond,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid auto save interval - too high",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    10 * time.Minute,
				Theme:               "dark",
			},
			wantErr: true,
		},
		{
			name: "invalid theme",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid light theme",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "light",
			},
			wantErr: false,
		},
		{
			name: "valid system theme",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "system",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidateWithDefaults(t *testing.T) {
	tests := []struct {
		name              string
		cfg               *Config
		expectedMaxHistory int
		expectedTheme     string
	}{
		{
			name: "fix invalid max history",
			cfg: &Config{
				MaxHistoryItems:     0,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			expectedMaxHistory: 1000,
			expectedTheme:       "dark",
		},
		{
			name: "fix invalid theme",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  500 * time.Millisecond,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "",
			},
			expectedMaxHistory: 100,
			expectedTheme:       "dark",
		},
		{
			name: "fix invalid monitoring interval",
			cfg: &Config{
				MaxHistoryItems:     100,
				MonitoringInterval:  0,
				PauseDuration:       30 * time.Second,
				AutoSaveInterval:    5 * time.Second,
				Theme:               "dark",
			},
			expectedMaxHistory: 100,
			expectedTheme:       "dark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.ValidateWithDefaults()
			if tt.cfg.MaxHistoryItems != tt.expectedMaxHistory {
				t.Errorf("MaxHistoryItems = %d, want %d", tt.cfg.MaxHistoryItems, tt.expectedMaxHistory)
			}
			if tt.cfg.Theme != tt.expectedTheme {
				t.Errorf("Theme = %s, want %s", tt.cfg.Theme, tt.expectedTheme)
			}
		})
	}
}

func TestConfigManagerLoadSave(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	// Test saving config
	cm := &ConfigManager{
		configPath: configPath,
		config: &Config{
			MaxHistoryItems:     500,
			MonitoringInterval:  1 * time.Second,
			Theme:               "light",
		},
	}
	
	err := cm.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}
	
	// Test loading config
	cm2 := &ConfigManager{
		configPath: configPath,
		config:     &Config{},
	}
	
	err = cm2.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify loaded values
	if cm2.config.MaxHistoryItems != 500 {
		t.Errorf("MaxHistoryItems = %d, want 500", cm2.config.MaxHistoryItems)
	}
	if cm2.config.Theme != "light" {
		t.Errorf("Theme = %s, want light", cm2.config.Theme)
	}
}

func TestConfigManagerLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.json")
	
	cm := &ConfigManager{
		configPath: configPath,
		config:     DefaultConfig(),
	}
	
	err := cm.Load()
	if err != nil {
		t.Fatalf("Failed to load non-existent config: %v", err)
	}
	
	// Should have saved the default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Default config should have been saved")
	}
}

func TestConfigManagerGet(t *testing.T) {
	cfg := &Config{MaxHistoryItems: 123}
	cm := &ConfigManager{config: cfg}
	
	got := cm.Get()
	if got != cfg {
		t.Error("Get() should return the same config pointer")
	}
	if got.MaxHistoryItems != 123 {
		t.Errorf("MaxHistoryItems = %d, want 123", got.MaxHistoryItems)
	}
}

func TestConfigManagerUpdate(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	cm := &ConfigManager{
		configPath: configPath,
		config:     DefaultConfig(),
	}
	
	newCfg := &Config{
		MaxHistoryItems: 2000,
		Theme:            "light",
	}
	
	err := cm.Update(newCfg)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}
	
	if cm.config.MaxHistoryItems != 2000 {
		t.Errorf("MaxHistoryItems = %d, want 2000", cm.config.MaxHistoryItems)
	}
	
	// Verify it was saved to disk
	cm2 := &ConfigManager{
		configPath: configPath,
		config:     &Config{},
	}
	
	err = cm2.Load()
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}
	
	if cm2.config.MaxHistoryItems != 2000 {
		t.Errorf("Loaded MaxHistoryItems = %d, want 2000", cm2.config.MaxHistoryItems)
	}
}

func TestConfigManagerMigrate(t *testing.T) {
	cm := &ConfigManager{
		config: &Config{
			MaxHistoryItems:     0,
			MonitoringInterval:  0,
			PauseDuration:       0,
			AutoSaveInterval:    0,
			Theme:               "",
		},
	}
	
	cm.migrate()
	
	if cm.config.MaxHistoryItems != 1000 {
		t.Errorf("MaxHistoryItems = %d, want 1000", cm.config.MaxHistoryItems)
	}
	if cm.config.MonitoringInterval != 500*time.Millisecond {
		t.Errorf("MonitoringInterval = %v, want 500ms", cm.config.MonitoringInterval)
	}
	if cm.config.PauseDuration != 30*time.Second {
		t.Errorf("PauseDuration = %v, want 30s", cm.config.PauseDuration)
	}
	if cm.config.AutoSaveInterval != 5*time.Second {
		t.Errorf("AutoSaveInterval = %v, want 5s", cm.config.AutoSaveInterval)
	}
	if cm.config.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", cm.config.Theme)
	}
}

func TestConfigManagerGetConfigPath(t *testing.T) {
	cm := &ConfigManager{
		configPath: "/test/path/config.json",
	}
	
	path := cm.GetConfigPath()
	if path != "/test/path/config.json" {
		t.Errorf("GetConfigPath() = %s, want /test/path/config.json", path)
	}
}
