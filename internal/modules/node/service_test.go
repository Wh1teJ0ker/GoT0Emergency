package node

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPlugins(t *testing.T) {
	// Setup temporary directory structure
	tmpDir := t.TempDir()
	
	// Create modules directory
	modulesDir := filepath.Join(tmpDir, "cmd/node/modules")
	err := os.MkdirAll(modulesDir, 0755)
	if err != nil {
		t.Fatalf("failed to create modules directory: %v", err)
	}

	// Create a test plugin directory
	pluginName := "testplugin"
	pluginDir := filepath.Join(modulesDir, pluginName)
	err = os.Mkdir(pluginDir, 0755)
	if err != nil {
		t.Fatalf("failed to create plugin directory: %v", err)
	}

	// Initialize service with project root set to temp dir
	svc := NewService()
	svc.SetProjectRoot(tmpDir)

	// Test GetPlugins
	plugins, err := svc.GetPlugins()
	if err != nil {
		t.Fatalf("GetPlugins failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
	if len(plugins) > 0 && plugins[0].Name != pluginName {
		t.Errorf("expected plugin name %s, got %s", pluginName, plugins[0].Name)
	}
}
