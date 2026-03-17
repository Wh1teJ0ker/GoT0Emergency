// Package log provides centralized logging functionality for the application
// Uses Go's standard slog package with JSON output
// Logs are written to dated files under data/logs/
package log

import (
	"GoT0Emergency/internal/pkg/path"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logger *slog.Logger

// Errorf wraps fmt.Errorf for consistent error formatting
func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// Sprintf wraps fmt.Sprintf for string formatting
func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Init initializes the logging system
// Creates a JSON log file for the current date under data/logs/
func Init() error {
	logFile := GetLogPath()
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(file, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)

	return nil
}

// Info logs an informational message with key-value pairs
func Info(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// Error logs an error message with key-value pairs
func Error(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

// Debug logs a debug message with key-value pairs
func Debug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}

// With returns a logger with the given context
func With(args ...any) *slog.Logger {
	if logger != nil {
		return logger.With(args...)
	}
	return slog.Default()
}

// GetLogPath returns the log file path for the current date
func GetLogPath() string {
	today := time.Now().Format("2006-01-02")
	return filepath.Join(path.GetLogDir(), fmt.Sprintf("app-%s.log", today))
}

// GetLogFiles returns a list of available log files (dates)
// Returns: slice of date strings extracted from filenames
func GetLogFiles() ([]string, error) {
	logDir := path.GetLogDir()
	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	var dates []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "app-") && strings.HasSuffix(file.Name(), ".log") {
			// Extract date: app-2023-10-27.log -> 2023-10-27
			name := file.Name()
			date := strings.TrimSuffix(strings.TrimPrefix(name, "app-"), ".log")
			dates = append(dates, date)
		}
	}
	return dates, nil
}

// ReadLogsByDate reads log entries for a specific date
// date: date string in format YYYY-MM-DD
// limit: maximum number of lines to return (0 for all)
// Returns: slice of log lines
func ReadLogsByDate(date string, limit int) ([]string, error) {
	logPath := filepath.Join(path.GetLogDir(), fmt.Sprintf("app-%s.log", date))
	content, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	// This is a simple implementation. For large files, we should read from the end.
	// But for now, reading all and taking last N lines is fine for MVP.
	sContent := string(content)
	lines := strings.Split(sContent, "\n")

	// Remove empty lines
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	if len(result) > limit && limit > 0 {
		return result[len(result)-limit:], nil
	}
	return result, nil
}

// ReadLogs reads log entries for the current date
// limit: maximum number of lines to return (0 for all)
func ReadLogs(limit int) ([]string, error) {
	today := time.Now().Format("2006-01-02")
	return ReadLogsByDate(today, limit)
}

// ClearLogs truncates the current log file to zero size
func ClearLogs() error {
	logPath := GetLogPath()
	return os.Truncate(logPath, 0)
}
