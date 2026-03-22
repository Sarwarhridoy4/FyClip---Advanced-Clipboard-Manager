// File: internal/config/validation.go
package config

import (
	"errors"
	"time"
)

var (
	// ErrInvalidMaxHistory is returned when max history is invalid
	ErrInvalidMaxHistory = errors.New("max_history_items must be between 1 and 100000")
	// ErrInvalidMonitoringInterval is returned when monitoring interval is invalid
	ErrInvalidMonitoringInterval = errors.New("monitoring_interval must be between 100ms and 10s")
	// ErrInvalidPauseDuration is returned when pause duration is invalid
	ErrInvalidPauseDuration = errors.New("pause_duration must be between 1s and 24h")
	// ErrInvalidAutoSaveInterval is returned when auto save interval is invalid
	ErrInvalidAutoSaveInterval = errors.New("auto_save_interval must be between 1s and 5m")
	// ErrInvalidTheme is returned when theme is invalid
	ErrInvalidTheme = errors.New("theme must be 'dark', 'light', or 'system'")
)

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate MaxHistoryItems
	if c.MaxHistoryItems < 1 || c.MaxHistoryItems > 100000 {
		return ErrInvalidMaxHistory
	}

	// Validate MonitoringInterval
	if c.MonitoringInterval < 100*time.Millisecond || c.MonitoringInterval > 10*time.Second {
		return ErrInvalidMonitoringInterval
	}

	// Validate PauseDuration
	if c.PauseDuration < 1*time.Second || c.PauseDuration > 24*time.Hour {
		return ErrInvalidPauseDuration
	}

	// Validate AutoSaveInterval
	if c.AutoSaveInterval < 1*time.Second || c.AutoSaveInterval > 5*time.Minute {
		return ErrInvalidAutoSaveInterval
	}

	// Validate Theme
	validThemes := map[string]bool{
		"dark":   true,
		"light":  true,
		"system": true,
	}
	if !validThemes[c.Theme] {
		return ErrInvalidTheme
	}

	return nil
}

// ValidateWithDefaults validates the configuration and applies defaults for any invalid values
func (c *Config) ValidateWithDefaults() {
	// Apply defaults for invalid values
	if c.MaxHistoryItems < 1 || c.MaxHistoryItems > 100000 {
		c.MaxHistoryItems = 1000
	}

	if c.MonitoringInterval < 100*time.Millisecond || c.MonitoringInterval > 10*time.Second {
		c.MonitoringInterval = 500 * time.Millisecond
	}

	if c.PauseDuration < 1*time.Second || c.PauseDuration > 24*time.Hour {
		c.PauseDuration = 30 * time.Second
	}

	if c.AutoSaveInterval < 1*time.Second || c.AutoSaveInterval > 5*time.Minute {
		c.AutoSaveInterval = 5 * time.Second
	}

	if c.Theme == "" {
		c.Theme = "dark"
	}
}
