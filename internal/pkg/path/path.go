// Package path provides centralized path management for the application
// Handles data directory, database path, log directory, and other runtime paths
package path

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	rootDir   string      // Project root directory
	dataDir   string      // Data directory (data/)
	dbDir     string      // Database directory (data/db/)
	dbPath    string      // Database file path
	logDir    string      // Log directory (data/logs/)
	sshDir    string      // SSH keys directory (data/ssh/)
	tmpDir    string      // Temporary directory (data/tmp/)
	nodeDir   string      // Node agent builds directory (data/nodes/)
	pathMutex sync.RWMutex // Mutex for thread-safe path operations
)

// Init initializes all runtime directories
// Creates necessary directory structure under the data directory
func Init() error {
	var err error
	rootDir, err = os.Getwd()
	if err != nil {
		return err
	}

	dataDir = filepath.Join(rootDir, "data")
	dbDir = filepath.Join(dataDir, "db")
	logDir = filepath.Join(dataDir, "logs")
	sshDir = filepath.Join(dataDir, "ssh")
	tmpDir = filepath.Join(dataDir, "tmp")
	nodeDir = filepath.Join(dataDir, "nodes")

	dirs := []string{dataDir, dbDir, logDir, sshDir, tmpDir, nodeDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	// Set default db path
	pathMutex.Lock()
	dbPath = filepath.Join(dbDir, "app.db")
	pathMutex.Unlock()

	return nil
}

// GetDataDir returns the data directory path
func GetDataDir() string {
	return dataDir
}

// GetDBPath returns the database file path
func GetDBPath() string {
	pathMutex.RLock()
	defer pathMutex.RUnlock()
	return dbPath
}

// SetDBPath sets a custom database path
// newPath: the new database file path
// Returns: error if failed to create the directory
func SetDBPath(newPath string) error {
	pathMutex.Lock()
	defer pathMutex.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dbPath = newPath
	return nil
}

// GetLogDir returns the log directory path
func GetLogDir() string {
	return logDir
}

// GetSSHDir returns the SSH keys directory path
func GetSSHDir() string {
	return sshDir
}

// GetTmpDir returns the temporary directory path
func GetTmpDir() string {
	return tmpDir
}

// GetNodeDir returns the Node agent builds directory path
func GetNodeDir() string {
	return nodeDir
}
