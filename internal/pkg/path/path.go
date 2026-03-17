package path

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	rootDir   string
	dataDir   string
	dbDir     string
	dbPath    string
	logDir    string
	sshDir    string
	tmpDir    string
	nodeDir   string
	pathMutex sync.RWMutex
)

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

func GetDataDir() string {
	return dataDir
}

func GetDBPath() string {
	pathMutex.RLock()
	defer pathMutex.RUnlock()
	return dbPath
}

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

func GetLogDir() string {
	return logDir
}

func GetSSHDir() string {
	return sshDir
}

func GetTmpDir() string {
	return tmpDir
}

func GetNodeDir() string {
	return nodeDir
}
