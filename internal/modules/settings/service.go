// Package settings provides application settings management functionality
// Handles storing and retrieving configuration settings in SQLite database
package settings

import (
	"database/sql"
	"strconv"

	"GoT0Emergency/internal/infra/db"
	"GoT0Emergency/internal/pkg/log"
)

const (
	// KeyRetentionHours is the settings key for log retention period in hours
	KeyRetentionHours = "retention_hours"
	// DefaultRetentionHours is the default log retention period (24 hours)
	DefaultRetentionHours = 24
)

// Service provides settings management operations
type Service struct{}

// NewService creates a new settings service instance
func NewService() *Service {
	return &Service{}
}

// Get retrieves a setting value by key
// key: the setting key
// Returns: setting value (empty string if not found) and error if query fails
func (s *Service) Get(key string) (string, error) {
	var value string
	err := db.DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", log.Errorf("failed to get setting %s: %w", key, err)
	}
	return value, nil
}

// Set sets a setting value
// key: the setting key
// value: the setting value
// Returns: error if insert fails
func (s *Service) Set(key, value string) error {
	_, err := db.DB.Exec("INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)", key, value)
	if err != nil {
		return log.Errorf("failed to set setting %s: %w", key, err)
	}
	return nil
}

// GetRetentionHours retrieves the log retention period in hours
// Returns: retention hours (defaults to DefaultRetentionHours if not set or invalid)
func (s *Service) GetRetentionHours() int {
	val, err := s.Get(KeyRetentionHours)
	if err != nil {
		log.Error("Failed to get retention hours", "err", err)
		return DefaultRetentionHours
	}
	if val == "" {
		return DefaultRetentionHours
	}
	hours, err := strconv.Atoi(val)
	if err != nil {
		log.Error("Invalid retention hours value", "val", val, "err", err)
		return DefaultRetentionHours
	}
	return hours
}

// SetRetentionHours sets the log retention period in hours
// hours: retention period in hours (must be positive)
// Returns: error if hours is not positive or if save fails
func (s *Service) SetRetentionHours(hours int) error {
	if hours <= 0 {
		return log.Errorf("retention hours must be positive")
	}
	return s.Set(KeyRetentionHours, strconv.Itoa(hours))
}
