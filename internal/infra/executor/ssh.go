// Package executor provides command execution interface and implementations
// Supports local execution and SSH remote execution modes
package executor

import (
	"GoT0Emergency/internal/pkg/log"
	"bytes"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHExecutor executes commands on a remote host via SSH
type SSHExecutor struct {
	client *ssh.Client // SSH client connection
}

// NewSSHExecutor creates a new SSH executor instance
// client: established SSH client connection
func NewSSHExecutor(client *ssh.Client) *SSHExecutor {
	return &SSHExecutor{client: client}
}

// Exec executes a command on the remote host via SSH
// cmdStr: command string to execute
// Returns: command stdout and error if execution fails
func (e *SSHExecutor) Exec(cmdStr string) (string, error) {
	if e.client == nil {
		return "", log.Errorf("ssh client is nil")
	}

	session, err := e.client.NewSession()
	if err != nil {
		return "", log.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmdStr)
	if err != nil {
		return "", log.Errorf("ssh exec failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// ConnectSSH creates a new SSH client connection
// ip: remote host IP address or hostname
// port: SSH port number
// user: username for authentication
// authMethods: list of SSH authentication methods
// Returns: SSH client instance and error if connection fails
// Note: Uses InsecureIgnoreHostKey() for MVP - consider adding host key verification for production
func ConnectSSH(ip string, port int, user string, authMethods []ssh.AuthMethod) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // MVP: Ignore host key verification
		Timeout:         10 * time.Second,            // Increased timeout for slow networks
		// Add legacy algorithms for compatibility with older servers
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoDSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
		// Keepalive configuration to maintain stable connections
		ClientVersion: "SSH-2.0-GoT0Emergency",
	}

	addr := log.Sprintf("%s:%d", ip, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	// Start keepalive goroutine
	go keepAlive(client, ip, port)

	return client, nil
}

// keepAlive sends keepalive packets every 30 seconds to maintain the SSH connection
// client: SSH client connection
// ip, port: connection target for logging purposes
func keepAlive(client *ssh.Client, ip string, port int) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		_, _, err := client.SendRequest("keepalive@got0emergency", true, nil)
		if err != nil {
			log.Debug("Keepalive failed, connection may be lost", "ip", ip, "port", port, "error", err)
			return
		}
	}
}
