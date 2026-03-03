package monitor

import (
	"encoding/json"
	"testing"
)

func TestCheckLocal(t *testing.T) {
	// Create a service with nil dependencies since checkLocal doesn't use them
	svc := &Service{}

	// ID 0 triggers local check
	status, err := svc.CheckHost(0)
	if err != nil {
		t.Fatalf("CheckHost(0) failed: %v", err)
	}
	if status == nil {
		t.Fatal("CheckHost(0) returned nil status")
	}

	// Test System Info
	if status.System.Hostname == "" {
		t.Error("Hostname is empty")
	}
	if status.System.OS == "" {
		t.Error("OS is empty")
	}

	// Test CPU
	if status.CPU.CoresLogical == 0 {
		t.Log("Logical Cores is 0 (might be possible in some virtual envs but unlikely)")
	}

	// Test Memory
	if status.Memory.Total == 0 {
		t.Error("Memory Total is 0")
	}

	// Test Disk
	// Note: In some CI/test environments, disk partitions might be empty or restricted
	if len(status.Disk.Partitions) > 0 {
		t.Logf("Found %d partitions", len(status.Disk.Partitions))
	} else {
		t.Log("No disk partitions found (might be expected in restricted env)")
	}

	// Print JSON for visual verification
	b, _ := json.MarshalIndent(status, "", "  ")
	t.Logf("Local Status:\n%s", string(b))
}
