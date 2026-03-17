// Package session provides SSH session management functionality
// Handles connection pooling, file transfer (SFTP), and remote port forwarding
package session

import (
	"io"
	"os"
	"sync"
	"time"

	"GoT0Emergency/internal/infra/executor"
	"GoT0Emergency/internal/modules/host"
	"GoT0Emergency/internal/pkg/log"

	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SessionManager manages SSH connections to remote hosts
// SessionManager manages SSH connections to remote hosts
type SessionManager struct {
	mu              sync.RWMutex
	connections     map[int64]*ssh.Client // Active SSH connections by host ID
	remoteListeners map[int64]net.Listener // Remote port listeners by host ID
	hostService     *host.Service         // Host service for retrieving host info
}

// FileInfo represents a remote file entry
type FileInfo struct {
	Name    string `json:"name"`    // File or directory name
	Size    int64  `json:"size"`    // File size in bytes
	IsDir   bool   `json:"is_dir"`  // True if directory
	ModTime string `json:"mod_time"` // Last modification time (RFC3339 format)
}

// NewSessionManager creates a new session manager instance
func NewSessionManager(hostService *host.Service) *SessionManager {
	return &SessionManager{
		connections: make(map[int64]*ssh.Client),
		remoteListeners: make(map[int64]net.Listener),
		hostService: hostService,
	}
}

// Connect establishes an SSH connection to a host
// hostID: the host unique identifier
// Returns: error if connection fails
func (sm *SessionManager) Connect(hostID int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.connections[hostID]; ok {
		return nil // Already connected
	}

	h, err := sm.hostService.GetHost(hostID)
	if err != nil {
		return err
	}

	log.Info("Connecting to host via SSH", "host", h.Name, "ip", h.IP, "port", h.Port)

	var authMethods []ssh.AuthMethod
	if h.AuthType == "password" {
		authMethods = append(authMethods, ssh.Password(h.Password))
		// Also try Keyboard Interactive as fallback
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			answers = make([]string, len(questions))
			for i := range questions {
				answers[i] = h.Password
			}
			return answers, nil
		}))
	} else if h.AuthType == "key" {
		key, err := os.ReadFile(h.KeyPath)
		if err != nil {
			return log.Errorf("unable to read private key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return log.Errorf("unable to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		authMethods = append(authMethods, ssh.Password(h.Password))
	}

	sshPort := h.Port
	if sshPort == 3389 {
		sshPort = 22
	}

	client, err := executor.ConnectSSH(h.IP, sshPort, h.Username, authMethods)
	if err != nil {
		log.Error("Failed to connect to host", "host", h.Name, "err", err)
		return err
	}

	sm.connections[hostID] = client
	_ = sm.hostService.UpdateLastConnected(hostID)
	log.Info("Connected to host successfully", "host", h.Name)
	return nil
}

// ListFiles lists files in a remote directory
// hostID: the host unique identifier
// remotePath: the remote directory path
// Returns: slice of FileInfo and error if operation fails
func (sm *SessionManager) ListFiles(hostID int64, remotePath string) ([]FileInfo, error) {
	client, err := sm.GetClient(hostID)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, log.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	var result []FileInfo
	// Add parent directory ".."
	result = append(result, FileInfo{
		Name:    "..",
		IsDir:   true,
		ModTime: "",
	})

	for _, f := range files {
		result = append(result, FileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			IsDir:   f.IsDir(),
			ModTime: f.ModTime().Format(time.RFC3339),
		})
	}
	return result, nil
}

// GetExecutor returns an SSH executor for a connected host
// hostID: the host unique identifier
// Returns: Executor interface and error if not connected
func (sm *SessionManager) GetExecutor(hostID int64) (executor.Executor, error) {
	client, err := sm.GetClient(hostID)
	if err != nil {
		return nil, err
	}
	return executor.NewSSHExecutor(client), nil
}

// GetClient returns the SSH client for a host
// hostID: the host unique identifier
// Returns: pointer to ssh.Client and error if retrieval fails
// Note: Will attempt auto-connect if not already connected
func (sm *SessionManager) GetClient(hostID int64) (*ssh.Client, error) {
	sm.mu.RLock()
	client, ok := sm.connections[hostID]
	sm.mu.RUnlock()

	if ok {
		log.Info("Host already connected, returning existing connection", "host_id", hostID)
		return client, nil
	}

	// Auto-connect if not connected
	log.Info("Host not connected, attempting auto-connect", "host_id", hostID)
	if err := sm.Connect(hostID); err != nil {
		log.Error("Auto-connect failed", "host_id", hostID, "error", err)
		return nil, err
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	client, ok = sm.connections[hostID]
	if !ok {
		return nil, log.Errorf("failed to retrieve connection after connect for host %d", hostID)
	}
	return client, nil
}

// getClientUnlocked returns the SSH client without locking (internal use)
// hostID: the host unique identifier
// Returns: pointer to ssh.Client and error if not connected
// Note: Does NOT auto-connect; caller must hold the mutex
func (sm *SessionManager) getClientUnlocked(hostID int64) (*ssh.Client, error) {
	client, ok := sm.connections[hostID]
	if !ok {
		// This is not an auto-connect path.
		return nil, log.Errorf("host %d not connected (cannot auto-connect within lock)", hostID)
	}
	return client, nil
}

// Close closes an SSH connection and cleans up associated resources
// hostID: the host unique identifier
// Returns: always nil (for API consistency)
func (sm *SessionManager) Close(hostID int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if client, ok := sm.connections[hostID]; ok {
		client.Close()
		delete(sm.connections, hostID)
	}
	if listener, ok := sm.remoteListeners[hostID]; ok {
		listener.Close()
		delete(sm.remoteListeners, hostID)
	}
	return nil
}

// ForwardRemotePort sets up remote port forwarding
// Forwards traffic from remotePort on the remote host to localPort on the local machine
// hostID: the host unique identifier
// remotePort: the port to listen on the remote host
// localPort: the local port to forward to
// Returns: error if setup fails
func (sm *SessionManager) ForwardRemotePort(hostID int64, remotePort, localPort int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.remoteListeners[hostID]; ok {
		log.Info("Remote port already forwarded for this host", "host_id", hostID, "port", remotePort)
		return nil // Already forwarding
	}

	client, err := sm.getClientUnlocked(hostID)
	if err != nil {
		return err
	}

	// Listen on remote port
	// 0.0.0.0 to listen on all interfaces on the remote host
	listener, err := client.Listen("tcp", log.Sprintf("0.0.0.0:%d", remotePort))
	if err != nil {
		return log.Errorf("failed to listen on remote port %d: %w", remotePort, err)
	}

	sm.remoteListeners[hostID] = listener
	log.Info("Started remote port forwarding", "host_id", hostID, "remote_port", remotePort, "local_port", localPort)

	go func() {
		defer func() {
			sm.mu.Lock()
			listener.Close()
			delete(sm.remoteListeners, hostID)
			sm.mu.Unlock()
			log.Info("Stopped remote port forwarding due to accept error", "host_id", hostID)
		}()

		for {
			remoteConn, err := listener.Accept()
			if err != nil {
				// This error is expected when the listener is closed, so we just exit the loop.
				return
			}

			localConn, err := net.Dial("tcp", log.Sprintf("localhost:%d", localPort))
			if err != nil {
				log.Error("Failed to dial local port for forwarding", "err", err)
				remoteConn.Close()
				continue
			}

			// Bidirectional copy
			go func() {
				defer remoteConn.Close()
				defer localConn.Close()
				io.Copy(remoteConn, localConn)
			}()
			go func() {
				defer remoteConn.Close()
				defer localConn.Close()
				io.Copy(localConn, remoteConn)
			}()
		}
	}()

	return nil
}

// IsConnected checks if a host is currently connected
// hostID: the host unique identifier
// Returns: true if connected, false otherwise
func (sm *SessionManager) IsConnected(hostID int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.connections[hostID]
	return ok
}

// UploadFile uploads a file from local to remote host via SFTP
// hostID: the host unique identifier
// localPath: path to the local file
// remotePath: destination path on the remote host
// resume: if true, resumes interrupted upload from checkpoint
// Returns: error if upload fails
func (sm *SessionManager) UploadFile(hostID int64, localPath, remotePath string, resume bool) error {
	client, err := sm.GetClient(hostID)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return log.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return log.Errorf("failed to open local file: %w", err)
	}
	defer srcFile.Close()

	var dstFile *sftp.File
	var offset int64 = 0

	if resume {
		info, err := sftpClient.Stat(remotePath)
		if err == nil {
			offset = info.Size()
			dstFile, err = sftpClient.OpenFile(remotePath, os.O_WRONLY|os.O_APPEND)
		} else {
			// If file doesn't exist, create it
			dstFile, err = sftpClient.Create(remotePath)
		}
	} else {
		dstFile, err = sftpClient.Create(remotePath)
	}

	if err != nil {
		return log.Errorf("failed to open/create remote file: %w", err)
	}
	defer dstFile.Close()

	if offset > 0 {
		if _, err := srcFile.Seek(offset, 0); err != nil {
			return log.Errorf("failed to seek local file: %w", err)
		}
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return log.Errorf("failed to copy file: %w", err)
	}

	// Only chmod if it's a new file or we want to ensure permissions
	if !resume || offset == 0 {
		if err := sftpClient.Chmod(remotePath, 0755); err != nil {
			return log.Errorf("failed to chmod remote file: %w", err)
		}
	}

	return nil
}

// RemoveFile deletes a file or directory on the remote host
// hostID: the host unique identifier
// remotePath: path to the remote file or directory
// Returns: error if deletion fails
func (sm *SessionManager) RemoveFile(hostID int64, remotePath string) error {
	executor, err := sm.GetExecutor(hostID)
	if err != nil {
		return err
	}

	// Use rm -rf to force delete files or directories
	// Be careful with quoting to avoid command injection if possible, though path is usually safe-ish here
	// Better to wrap path in quotes
	cmd := log.Sprintf("rm -rf \"%s\"", remotePath)
	_, err = executor.Exec(cmd)
	if err != nil {
		return log.Errorf("failed to remove file/dir: %w", err)
	}
	return nil
}

// DownloadFile downloads a file from remote host to local via SFTP
// hostID: the host unique identifier
// remotePath: path to the remote file
// localPath: destination path on the local machine
// resume: if true, resumes interrupted download from checkpoint
// Returns: error if download fails
func (sm *SessionManager) DownloadFile(hostID int64, remotePath, localPath string, resume bool) error {
	client, err := sm.GetClient(hostID)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return log.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return log.Errorf("failed to open remote file: %w", err)
	}
	defer srcFile.Close()

	var dstFile *os.File
	var offset int64 = 0

	if resume {
		info, err := os.Stat(localPath)
		if err == nil {
			offset = info.Size()
			dstFile, err = os.OpenFile(localPath, os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			dstFile, err = os.Create(localPath)
		}
	} else {
		dstFile, err = os.Create(localPath)
	}

	if err != nil {
		return log.Errorf("failed to create/open local file: %w", err)
	}
	defer dstFile.Close()

	if offset > 0 {
		if _, err := srcFile.Seek(offset, 0); err != nil {
			return log.Errorf("failed to seek remote file: %w", err)
		}
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return log.Errorf("failed to copy file: %w", err)
	}

	return nil
}
