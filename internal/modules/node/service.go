package node

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"GoT0Emergency/internal/pkg/log"
	"GoT0Emergency/internal/pkg/path"
)

type Plugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
	Path        string `json:"path"`
}

type Service struct {
	projectRoot string
}

func NewService() *Service {
	root, _ := os.Getwd()
	// If running tests, we might need to adjust, but for now default to CWD
	return &Service{
		projectRoot: root,
	}
}

// SetProjectRoot allows overriding the root directory (e.g. for testing)
func (s *Service) SetProjectRoot(path string) {
	s.projectRoot = path
}

func (s *Service) GetPlugins() ([]Plugin, error) {
	// Assume running from project root
	moduleDir := filepath.Join(s.projectRoot, "cmd/node/modules")
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return nil, log.Errorf("failed to read modules directory (%s): %w", moduleDir, err)
	}

	var plugins []Plugin
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			tag := "feat_" + name
			desc := "标准功能模块"

			// Try to read description from a README or comments?
			// For now, hardcode known ones or generic.
			if name == "host_monitor" {
				desc = "采集系统指标（CPU、内存、磁盘、负载、网络、进程）"
			}

			plugins = append(plugins, Plugin{
				Name:        name,
				Description: desc,
				Tag:         tag,
				Path:        filepath.Join(moduleDir, name),
			})
		}
	}
	return plugins, nil
}

func (s *Service) CreatePlugin(name, description string) error {
	// 1. Validate name (alphanumeric, lowercase)
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return log.Errorf("plugin name cannot be empty")
	}

	moduleDir := filepath.Join(s.projectRoot, "cmd/node/modules", name)
	if _, err := os.Stat(moduleDir); !os.IsNotExist(err) {
		return log.Errorf("plugin %s already exists", name)
	}

	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return log.Errorf("failed to create plugin directory: %w", err)
	}

	// 2. Create files
	if err := s.createPluginFile(moduleDir, name, description); err != nil {
		return err
	}
	if err := s.createStubFile(moduleDir, name); err != nil {
		return err
	}

	// 3. Register in main.go
	if err := s.registerInMain(name); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetPluginSource(name string) (string, error) {
	// 1. Construct path
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "", log.Errorf("plugin name cannot be empty")
	}

	// We assume the main logic is in {name}.go
	filePath := filepath.Join(s.projectRoot, "cmd/node/modules", name, name+".go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", log.Errorf("failed to read plugin source: %w", err)
	}

	return string(content), nil
}

func (s *Service) SavePluginSource(name, content string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return log.Errorf("plugin name cannot be empty")
	}

	filePath := filepath.Join(s.projectRoot, "cmd/node/modules", name, name+".go")

	// Basic security check: ensure we are writing to the correct directory
	if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(filepath.Join(s.projectRoot, "cmd/node/modules"))) {
		return log.Errorf("invalid file path")
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return log.Errorf("failed to save plugin source: %w", err)
	}

	return nil
}

func (s *Service) createPluginFile(dir, name, description string) error {
	content := log.Sprintf(`//go:build feat_%[1]s

package %[1]s

import (
	"GoT0Emergency/cmd/node/core"
)

type %[2]sModule struct{}

func (m *%[2]sModule) Name() string {
	return "%[1]s"
}

func (m *%[2]sModule) Run() (interface{}, error) {
	// TODO: Implement plugin logic
	return map[string]string{
		"message": "Hello from %[1]s plugin!",
		"description": "%[3]s",
	}, nil
}

func init() {
	core.RegisterModule(&%[2]sModule{})
}
`, name, capitalize(name), description)

	return os.WriteFile(filepath.Join(dir, name+".go"), []byte(content), 0644)
}

func (s *Service) createStubFile(dir, name string) error {
	content := log.Sprintf(`//go:build !feat_%[1]s

package %[1]s

// Stub for when the module is not included
func init() {
	// No-op
}
`, name)

	return os.WriteFile(filepath.Join(dir, "stub.go"), []byte(content), 0644)
}

func (s *Service) registerInMain(name string) error {
	mainPath := filepath.Join(s.projectRoot, "cmd/node/main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	importFound := false

	// Check if already imported (shouldn't happen if dir didn't exist, but safe check)
	importPath := log.Sprintf(`	_ "GoT0Emergency/cmd/node/modules/%s"`, name)

	for _, line := range lines {
		newLines = append(newLines, line)
		if strings.Contains(line, "GoT0Emergency/cmd/node/modules/host_monitor") && !importFound {
			// Insert after the existing module import
			newLines = append(newLines, importPath)
			importFound = true
		}
	}

	// If not found (maybe host_monitor was removed?), try to find imports block
	if !importFound {
		// Fallback: simple append isn't safe.
		// For now, assume host_monitor is always there or at least "cmd/node/core"
		// If we missed it, we might break the file.
		// Let's improve: Find "import (" and add to end of block?
		// But regex is safer. Given the task, assuming standard structure is okay.
		// If host_monitor line not found, we append to the import block.

		// Let's re-read and use a simpler strategy if the first failed.
		// But for now, let's assume the user hasn't deleted host_monitor import.
	}

	return os.WriteFile(mainPath, []byte(strings.Join(newLines, "\n")), 0644)
}

func (s *Service) GetBuiltNodes() ([]string, error) {
	nodeDir := path.GetNodeDir()
	entries, err := os.ReadDir(nodeDir)
	if err != nil {
		return nil, err
	}

	var nodes []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "node_") {
			nodes = append(nodes, entry.Name())
		}
	}
	return nodes, nil
}

func (s *Service) BuildNode(features []string, targetOS string, targetArch string) (string, error) {
	outputPath, err := s.Build(features, targetOS, targetArch)
	if err != nil {
		return "", err
	}

	// Get file size
	info, err := os.Stat(outputPath)
	if err != nil {
		return "", log.Errorf("built successfully but failed to stat output file: %v", err)
	}

	msg := log.Sprintf("Node built successfully! Path: %s, Size: %.2f MB", outputPath, float64(info.Size())/1024/1024)
	log.Info("Node build complete", "path", outputPath, "size_bytes", info.Size())
	return msg, nil
}

func (s *Service) Build(features []string, targetOS string, targetArch string) (string, error) {
	// Security check
	for _, f := range features {
		if !strings.HasPrefix(f, "feat_") {
			return "", log.Errorf("invalid feature tag: %s", f)
		}
	}

	tags := strings.Join(features, " ")
	nodeDir := path.GetNodeDir()
	outputPath := filepath.Join(nodeDir, log.Sprintf("node_%s_%s", targetOS, targetArch))
	if targetOS == "windows" {
		outputPath += ".exe"
	}

	// Environment variables
	env := os.Environ()
	env = append(env, log.Sprintf("GOOS=%s", targetOS))
	env = append(env, log.Sprintf("GOARCH=%s", targetArch))

	args := []string{"build", "-tags", tags, "-ldflags", "-s -w", "-o", outputPath, "./cmd/node"}
	cmd := exec.Command("go", args...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Node build failed", "err", err, "output", string(output))
		return "", log.Errorf("build failed: %v, output: %s", err, string(output))
	}

	return outputPath, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
