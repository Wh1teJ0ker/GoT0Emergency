package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	stdRuntime "runtime"
	"strings"
	"time"

	"GoT0Emergency/internal/infra/db"
	"GoT0Emergency/internal/infra/executor"
	"GoT0Emergency/internal/infra/session"
	"GoT0Emergency/internal/modules/host"
	"GoT0Emergency/internal/modules/monitor"
	"GoT0Emergency/internal/modules/node"
	"GoT0Emergency/internal/modules/settings"
	"GoT0Emergency/internal/modules/terminal"
	"GoT0Emergency/internal/modules/terminal/local"
	"GoT0Emergency/internal/modules/terminal/rdp"
	"GoT0Emergency/internal/modules/terminal/ssh"
	"GoT0Emergency/internal/pkg/log"
	"GoT0Emergency/internal/pkg/path"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	hostService     *host.Service
	sessionManager  *session.SessionManager
	localExecutor   *executor.LocalExecutor
	terminalManager *terminal.Manager
	monitorService  *monitor.Service
	nodeService     *node.Service
	settingsService *settings.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	hs := host.NewService()
	sm := session.NewSessionManager(hs)
	le := executor.NewLocalExecutor()
	set := settings.NewService()
	return &App{
		hostService:     hs,
		sessionManager:  sm,
		localExecutor:   le,
		terminalManager: terminal.NewManager(),
		monitorService:  monitor.NewService(sm, le, hs, set),
		nodeService:     node.NewService(),
		settingsService: set,
	}
}

// Settings Methods
func (a *App) GetRetentionHours() int {
	return a.settingsService.GetRetentionHours()
}

func (a *App) SetRetentionHours(hours int) error {
	return a.settingsService.SetRetentionHours(hours)
}

// Monitor Methods
func (a *App) CheckHostStatus(hostID int64) (*monitor.HostStatus, error) {
	// Bypass cache for explicit checks (e.g., from UI refresh)
	// We want fresh data when the user asks for it.
	return a.monitorService.CheckHost(hostID)
}

func (a *App) GetHostMetrics(hostID int64, durationStr string) ([]monitor.MetricPoint, error) {
	return a.monitorService.GetHostMetrics(hostID, durationStr)
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// Start monitor service background tasks
	a.monitorService.Start()
	// Start callback server
	go a.startCallbackServer(36911)
}

func (a *App) Shutdown(ctx context.Context) {
	if a.monitorService != nil {
		a.monitorService.Stop()
	}
}

func (a *App) startCallbackServer(port int) {
	log.Info("Starting callback server", "port", port)

	// Use a new ServeMux to avoid global state issues if called multiple times or conflicting
	mux := http.NewServeMux()
	mux.HandleFunc("/api/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Callback request received", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("Failed to read callback body", "err", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Info("Callback body read", "size", len(body), "raw", string(body))

		// Unmarshal
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			log.Error("Failed to parse node callback JSON", "err", err)
			w.WriteHeader(http.StatusOK) // Return OK to avoid retry storm
			return
		}

		// Log callback
		hostname, _ := data["hostname"].(string)
		log.Info("Received node callback", "remote_ip", r.RemoteAddr, "host", hostname, "data", data)

		// Add remote IP
		data["remote_ip"] = r.RemoteAddr

		// Emit event to frontend with data
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "node:callback", data)
			log.Info("Emitted callback event to frontend")
		} else {
			log.Error("App context is nil, cannot emit event")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    log.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Info("Callback server listening", "addr", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error("Failed to start callback server", "err", err)
	}
}

// DB Methods
func (a *App) InitDB() error {
	return db.Init()
}

func (a *App) GetDBPath() string {
	return path.GetDBPath()
}

// Log Methods
func (a *App) GetLogs(limit int) ([]string, error) {
	return log.ReadLogs(limit)
}

func (a *App) GetLogFiles() ([]string, error) {
	return log.GetLogFiles()
}

func (a *App) GetLogsByDate(date string, limit int) ([]string, error) {
	return log.ReadLogsByDate(date, limit)
}

func (a *App) ClearLogs() error {
	return log.ClearLogs()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return log.Sprintf("Hello %s, It's show time!", name)
}

// Host Methods
func (a *App) GetHosts() ([]host.Host, error) {
	return a.hostService.GetHosts()
}

func (a *App) GetHost(id int64) (*host.Host, error) {
	return a.hostService.GetHost(id)
}

func (a *App) CreateHost(h host.Host) error {
	return a.hostService.CreateHost(&h)
}

func (a *App) DeleteHost(id int64) error {
	return a.hostService.DeleteHost(id)
}

func (a *App) UpdateHost(h host.Host) error {
	return a.hostService.UpdateHost(&h)
}

// SSH Methods
func (a *App) ConnectSSH(hostID int64) error {
	return a.sessionManager.Connect(hostID)
}

func (a *App) CloseSSH(hostID int64) error {
	return a.sessionManager.Close(hostID)
}

func (a *App) IsConnected(hostID int64) bool {
	return a.sessionManager.IsConnected(hostID)
}

// RDP Methods
func (a *App) LaunchRDP(hostID int64, width, height int) error {
	hosts, err := a.hostService.GetHosts()
	if err != nil {
		return err
	}
	var targetHost host.Host
	found := false
	for _, h := range hosts {
		if h.ID == hostID {
			targetHost = h
			found = true
			break
		}
	}
	if !found {
		return log.Errorf("host not found")
	}

	config := rdp.RDPConfig{
		Host:     targetHost.IP,
		Port:     targetHost.Port,
		Username: targetHost.Username,
		Password: targetHost.Password,
		Width:    width,
		Height:   height,
	}
	// Default port for RDP is 3389, but if user set a custom port in Host, use it.
	// If port is 22 (SSH default), we should probably warn or default to 3389 if not specified?
	// But let's assume the user set the correct port for the RDP service if they are clicking "RDP".
	// Or we can check if port is 22 and default to 3389.
	if config.Port == 22 {
		config.Port = 3389
	}

	return rdp.LaunchRDP(config)
}

// Node Methods
func (a *App) GetBuiltNodes() ([]string, error) {
	return a.nodeService.GetBuiltNodes()
}

func (a *App) GetNodeModules() ([]node.Plugin, error) {
	return a.nodeService.GetPlugins()
}

func (a *App) CreatePlugin(name, description string) error {
	return a.nodeService.CreatePlugin(name, description)
}

func (a *App) GetPluginSource(name string) (string, error) {
	return a.nodeService.GetPluginSource(name)
}

func (a *App) SavePluginSource(name, content string) error {
	return a.nodeService.SavePluginSource(name, content)
}

func (a *App) BuildNode(features []string, targetOS string, targetArch string) (string, error) {
	return a.nodeService.BuildNode(features, targetOS, targetArch)
}

func (a *App) DeployNode(hostID int64, targetOS, targetArch string) (string, error) {
	log.Info("DeployNode called", "host_id", hostID, "os", targetOS, "arch", targetArch)

	// 1. Build Node (default features)
	features := []string{"feat_host_monitor"} // Default features
	log.Info("Building node...", "features", features)
	nodePath, err := a.nodeService.Build(features, targetOS, targetArch)
	if err != nil {
		log.Error("Node build failed", "error", err)
		return "", err
	}
	log.Info("Node built successfully", "path", nodePath)

	// 2. Connect first to get executor and determine remote path
	log.Info("Connecting to host...", "host_id", hostID)
	err = a.sessionManager.Connect(hostID)
	if err != nil {
		log.Error("Failed to connect to host", "error", err)
		return "", log.Errorf("failed to connect: %w", err)
	}
	log.Info("Connected to host successfully")

	executor, err := a.sessionManager.GetExecutor(hostID)
	if err != nil {
		log.Error("Failed to get executor", "error", err)
		return "", err
	}

	// Determine remote path - use home directory to avoid /tmp permission issues
	var remotePath string
	if targetOS == "windows" {
		remotePath = "C:\\Windows\\Temp\\got0_node.exe"
	} else {
		// Expand ~ to actual home directory
		output, err := executor.Exec("echo $HOME")
		if err != nil {
			log.Error("Failed to get home directory, using /tmp as fallback", "error", err)
			remotePath = "/tmp/got0_node"
		} else {
			homeDir := strings.TrimSpace(output)
			dirPath := homeDir + "/got0_node"
			log.Info("Using home directory for node", "home", homeDir, "dir", dirPath)

			// Remove existing directory first to ensure clean state
			_, _ = executor.Exec("rm -rf " + dirPath)
			log.Info("Removed existing directory if any", "dir", dirPath)

			// Create directory
			_, err = executor.Exec("mkdir -p " + dirPath)
			if err != nil {
				log.Error("Failed to create remote directory, using /tmp as fallback", "error", err)
				dirPath = "/tmp/got0_node"
				_, _ = executor.Exec("rm -rf " + dirPath)
				_, err = executor.Exec("mkdir -p " + dirPath)
			}
			remotePath = dirPath
		}
	}

	// For non-Windows, the remotePath is now a directory, append the filename
	if targetOS != "windows" {
		remotePath = remotePath + "/node"
	}

	log.Info("Uploading node...", "remote_path", remotePath)
	err = a.sessionManager.UploadFile(hostID, nodePath, remotePath, false)
	if err != nil {
		log.Error("Upload failed", "error", err)
		return "", log.Errorf("upload failed: %w", err)
	}
	log.Info("Node uploaded successfully", "remote_path", remotePath)

	// 3. Setup port forwarding BEFORE executing node
	log.Info("Setting up remote port forwarding", "host_id", hostID, "remote_port", 36911, "local_port", 36911)
	err = a.sessionManager.ForwardRemotePort(hostID, 36911, 36911)
	if err != nil {
		log.Error("Failed to setup remote port forwarding. Node will be unable to callback.", "host_id", hostID, "error", err)
		return "", log.Errorf("failed to setup remote port forwarding: %w", err)
	}
	log.Info("Remote port forwarding setup successfully")

	// 4. Execute Node (in background)
	var cmdStr string
	if targetOS == "windows" {
		cmdStr = log.Sprintf("powershell -Command \"Start-Process -FilePath '%s' -ArgumentList '-callback http://127.0.0.1:36911/api/callback' -WindowStyle Hidden -RedirectStandardOutput 'C:\\Windows\\Temp\\got0_node.log' -RedirectStandardError 'C:\\Windows\\Temp\\got0_node.err'\"", remotePath)
	} else {
		cmdStr = log.Sprintf("nohup %s -callback http://127.0.0.1:36911/api/callback > %s.log 2>&1 &", remotePath, remotePath)
	}

	log.Info("Executing node command", "cmd", cmdStr)
	output, err := executor.Exec(cmdStr)
	log.Info("Node execution result", "output", output, "error", err)

	if err != nil {
		log.Error("Failed to start node", "err", err)
	}

	// 5. Add a small delay to allow node to start and callback
	log.Info("Waiting for node to start and send callback...")
	time.Sleep(2 * time.Second)

	return "Node deployed and started successfully!", nil
}

func (a *App) RunNode(hostID int64) (string, error) {
	// 1. Determine Remote Path
	// Ideally we should know the OS/Arch of the host.
	// For now, we assume standard paths.
	// TODO: Store deployed node path in DB or check multiple paths.

	// Try Linux path first
	remotePath := "/tmp/got0_node"

	executor, err := a.sessionManager.GetExecutor(hostID)
	if err != nil {
		return "", err
	}

	// Check if file exists (simple check via ls)
	_, err = executor.Exec(log.Sprintf("ls %s", remotePath))
	if err != nil {
		// Try Windows path
		remotePath = "C:\\Windows\\Temp\\got0_node.exe"
		_, err = executor.Exec(log.Sprintf("dir %s", remotePath))
		if err != nil {
			return "", log.Errorf("node binary not found on remote host (checked /tmp/got0_node and C:\\Windows\\Temp\\got0_node.exe)")
		}
	}

	// 2. Execute Node Synchronously
	output, err := executor.Exec(remotePath)
	if err != nil {
		return "", log.Errorf("failed to run node: %w", err)
	}

	return output, nil
}

// File Transfer Methods
func (a *App) UploadFile(hostID int64, localPath, remotePath string, resume bool) error {
	return a.sessionManager.UploadFile(hostID, localPath, remotePath, resume)
}

func (a *App) DownloadFile(hostID int64, remotePath, localPath string, resume bool) error {
	return a.sessionManager.DownloadFile(hostID, remotePath, localPath, resume)
}

func (a *App) RemoveRemoteFile(hostID int64, remotePath string) error {
	return a.sessionManager.RemoveFile(hostID, remotePath)
}

func (a *App) ListRemoteFiles(hostID int64, remotePath string) ([]session.FileInfo, error) {
	return a.sessionManager.ListFiles(hostID, remotePath)
}

func (a *App) RDPOpen(hostID int64) error {
	// Get host IP
	hosts, err := a.hostService.GetHosts()
	if err != nil {
		return err
	}
	var targetHost *host.Host
	for _, h := range hosts {
		if int64(h.ID) == hostID {
			targetHost = &h
			break
		}
	}
	if targetHost == nil {
		return log.Errorf("host not found")
	}

	// Launch RDP client
	if stdRuntime.GOOS == "darwin" {
		return exec.Command("open", log.Sprintf("rdp://%s", targetHost.IP)).Start()
	} else if stdRuntime.GOOS == "windows" {
		return exec.Command("mstsc", "/v:"+targetHost.IP).Start()
	} else {
		// Linux: remmina? xfreerdp?
		return log.Errorf("RDP launch not supported on this OS: %s", stdRuntime.GOOS)
	}
}

// Execution
func (a *App) Exec(hostID int64, cmd string) (string, error) {
	var exec executor.Executor
	var err error

	if hostID == 0 {
		exec = a.localExecutor
	} else {
		exec, err = a.sessionManager.GetExecutor(hostID)
		if err != nil {
			return "", err
		}
	}

	return exec.Exec(cmd)
}

// Dialog Methods
func (a *App) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File",
	})
	return selection, err
}

func (a *App) SelectSaveFile() (string, error) {
	selection, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Save File",
	})
	return selection, err
}

// Terminal Methods
func (a *App) TerminalOpen(hostID int64, rows, cols int) (string, error) {
	id := terminal.GenerateID()
	var term terminal.Terminal
	var err error

	log.Info("Opening terminal", "host_id", hostID, "terminal_id", id)

	if hostID == 0 {
		term, err = local.NewLocalTerminal(a.ctx, id)
	} else {
		log.Info("Getting SSH client for terminal", "host_id", hostID)
		client, err := a.sessionManager.GetClient(hostID)
		if err != nil {
			log.Error("Failed to get SSH client", "host_id", hostID, "error", err)
			return "", err
		}
		term, err = ssh.NewSSHTerminal(a.ctx, client, id, rows, cols)
	}

	if err != nil {
		log.Error("Failed to create terminal", "host_id", hostID, "error", err)
		return "", err
	}

	a.terminalManager.Add(id, term)
	log.Info("Terminal created successfully", "terminal_id", id, "host_id", hostID)
	return id, nil
}

func (a *App) TerminalWrite(id string, data string) error {
	t, ok := a.terminalManager.Get(id)
	if !ok {
		return log.Errorf("terminal not found")
	}
	_, err := t.Write([]byte(data))
	return err
}

func (a *App) TerminalResize(id string, rows, cols int) error {
	t, ok := a.terminalManager.Get(id)
	if !ok {
		return log.Errorf("terminal not found")
	}
	return t.Resize(rows, cols)
}

func (a *App) TerminalClose(id string) error {
	a.terminalManager.Remove(id)
	return nil
}

// OpenURL opens the given URL in the system default browser
func (a *App) OpenURL(url string) error {
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}
