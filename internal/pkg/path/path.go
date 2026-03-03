package path

import (
	"os"
	"path/filepath"
)

var (
	rootDir string
	dataDir string
	dbDir   string
	logDir  string
	sshDir  string
	tmpDir  string
	nodeDir string
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
	return nil
}

func GetDataDir() string {
	return dataDir
}

func GetDBPath() string {
	return filepath.Join(dbDir, "app.db")
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
