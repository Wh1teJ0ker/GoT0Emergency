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

type SessionManager struct {
	mu                sync.RWMutex
	connections       map[int64]*ssh.Client
	remoteListeners   map[int64]net.Listener
	hostService       *host.Service
}

type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

func NewSessionManager(hostService *host.Service) *SessionManager {
	return &SessionManager{
		connections: make(map[int64]*ssh.Client),
		remoteListeners: make(map[int64]net.Listener),
		hostService: hostService,
	}
}

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

func (sm *SessionManager) GetExecutor(hostID int64) (executor.Executor, error) {
	client, err := sm.GetClient(hostID)
	if err != nil {
		return nil, err
	}
	return executor.NewSSHExecutor(client), nil
}

func (sm *SessionManager) GetClient(hostID int64) (*ssh.Client, error) {
	sm.mu.RLock()
	client, ok := sm.connections[hostID]
	sm.mu.RUnlock()

	if ok {
		return client, nil
	}

	// Auto-connect if not connected
	log.Info("Host not connected, attempting auto-connect", "host_id", hostID)
	if err := sm.Connect(hostID); err != nil {
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

// getClientUnlocked is for internal use when the session manager's mutex is already held.
// It does NOT auto-connect and assumes the caller has locked the mutex.
func (sm *SessionManager) getClientUnlocked(hostID int64) (*ssh.Client, error) {
	client, ok := sm.connections[hostID]
	if !ok {
		// This is not an auto-connect path.
		return nil, log.Errorf("host %d not connected (cannot auto-connect within lock)", hostID)
	}
	return client, nil
}

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

func (sm *SessionManager) ForwardRemotePort(hostID int64, remotePort, localPort int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.remoteListeners[hostID]; ok {
		log.Info("Remote port already forwarded for this host", "host_id", hostID, "port", remotePort)
		return nil // Already forwarding
	}

	client, err := sm.getClientUnlocked(hostID)
	if err != nil {
		log.Error("Failed to get SSH client for port forwarding", "host_id", hostID, "error", err)
		return err
	}

	// Listen on remote port
	// Try binding to 127.0.0.1 first (works without GatewayPorts)
	// If node runs on same host, it can still connect to 127.0.0.1
	remoteAddr := log.Sprintf("127.0.0.1:%d", remotePort)
	log.Info("Attempting to listen on remote address", "address", remoteAddr)
	listener, err := client.Listen("tcp", remoteAddr)
	if err != nil {
		// Try fallback to 0.0.0.0 (requires GatewayPorts yes)
		log.Error("Failed to bind to 127.0.0.1, trying 0.0.0.0 (requires GatewayPorts)", "error", err)
		remoteAddr = log.Sprintf("0.0.0.0:%d", remotePort)
		listener, err = client.Listen("tcp", remoteAddr)
		if err != nil {
			log.Error("Failed to listen on remote port", "address", remoteAddr, "error", err)
			return log.Errorf("failed to listen on remote port %d: %w", remotePort, err)
		}
	}

	sm.remoteListeners[hostID] = listener
	log.Info("Started remote port forwarding", "host_id", hostID, "remote_addr", remoteAddr, "local_port", localPort)

	go func() {
		defer func() {
			sm.mu.Lock()
			listener.Close()
			delete(sm.remoteListeners, hostID)
			sm.mu.Unlock()
			log.Info("Stopped remote port forwarding", "host_id", hostID)
		}()

		for {
			remoteConn, err := listener.Accept()
			if err != nil {
				// This error is expected when the listener is closed, so we just exit the loop.
				log.Debug("Listener accept error, stopping forwarder", "host_id", hostID, "error", err)
				return
			}

			log.Info("Accepted remote connection, dialing local port", "remote_addr", remoteConn.RemoteAddr(), "local_port", localPort)
			localConn, err := net.Dial("tcp", log.Sprintf("localhost:%d", localPort))
			if err != nil {
				log.Error("Failed to dial local port for forwarding", "err", err)
				remoteConn.Close()
				continue
			}

			log.Info("Port forwarding established", "remote", remoteConn.RemoteAddr(), "local", localConn.LocalAddr())

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

func (sm *SessionManager) IsConnected(hostID int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.connections[hostID]
	return ok
}

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
