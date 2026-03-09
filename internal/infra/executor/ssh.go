package executor

import (
	"GoT0Emergency/internal/pkg/log"
	"bytes"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	client *ssh.Client
}

func NewSSHExecutor(client *ssh.Client) *SSHExecutor {
	return &SSHExecutor{client: client}
}

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

// Helper to create client (usually handled in session manager, but useful here for simple setup)
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
	}

	addr := log.Sprintf("%s:%d", ip, port)
	return ssh.Dial("tcp", addr, config)
}
