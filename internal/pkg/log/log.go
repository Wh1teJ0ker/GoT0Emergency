package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"GoT0Emergency/internal/pkg/path"
)

var logger *slog.Logger

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

func Info(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

func Debug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}

func With(args ...any) *slog.Logger {
	if logger != nil {
		return logger.With(args...)
	}
	return slog.Default()
}

func GetLogPath() string {
	today := time.Now().Format("2006-01-02")
	return filepath.Join(path.GetLogDir(), fmt.Sprintf("app-%s.log", today))
}

// GetLogFiles returns a list of available log files (dates)
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

func ReadLogs(limit int) ([]string, error) {
	today := time.Now().Format("2006-01-02")
	return ReadLogsByDate(today, limit)
}

func ClearLogs() error {
	logPath := GetLogPath()
	return os.Truncate(logPath, 0)
}
