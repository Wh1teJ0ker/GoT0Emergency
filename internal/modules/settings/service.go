package settings

import (
	"database/sql"
	"strconv"

	"GoT0Emergency/internal/infra/db"
	"GoT0Emergency/internal/pkg/log"
)

const KeyRetentionHours = "retention_hours"
const DefaultRetentionHours = 24

type Service struct{}

func NewService() *Service {
	return &Service{}
}

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

func (s *Service) Set(key, value string) error {
	_, err := db.DB.Exec("INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)", key, value)
	if err != nil {
		return log.Errorf("failed to set setting %s: %w", key, err)
	}
	return nil
}

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

func (s *Service) SetRetentionHours(hours int) error {
	if hours <= 0 {
		return log.Errorf("retention hours must be positive")
	}
	return s.Set(KeyRetentionHours, strconv.Itoa(hours))
}
