package monitor

import (
	"strings"
	"testing"
)

type MockExecutor struct {
	Outputs map[string]string
	Errors  map[string]error
}

func (m *MockExecutor) Exec(cmd string) (string, error) {
	// Exact match first
	if val, ok := m.Outputs[cmd]; ok {
		return val, nil
	}
	// Partial match
	for k, v := range m.Outputs {
		if strings.Contains(cmd, k) {
			return v, nil
		}
	}
	// Error handling
	if m.Errors != nil {
		if err, ok := m.Errors[cmd]; ok {
			return "", err
		}
	}
	return "", nil
}

func TestCheckWindowsRemote(t *testing.T) {
	mockExec := &MockExecutor{
		Outputs: map[string]string{
			"hostname":                                      "win-test-host",
			"echo %USERNAME%":                               "Administrator",
			"(Get-CimInstance Win32_Processor).Name":        "Intel Core i7",
			"(Get-CimInstance Win32_Processor).NumberOfCores": "4",
			"(Get-CimInstance Win32_Processor).NumberOfLogicalProcessors": "8",
			"(Get-CimInstance Win32_Processor).MaxClockSpeed": "3200",
			"Measure-Object -Property LoadPercentage":       "15.5",
			"Win32_OperatingSystem; Write-Output":           "16777216\r\n8388608", // Total, Free (CRLF)
			"LastBootUpTime":                                "1.5",                 // Uptime days
			"Get-PSDrive":                                   "C 50000000 50000000",
			"Win32_PerfRawData_PerfDisk_PhysicalDisk":       "1000 2000 10 20",
			"Get-NetAdapterStatistics":                      "Ethernet|10000|5000",
			"Get-NetIPAddress":                              "Ethernet|192.168.1.10",
			"(Get-NetTCPConnection).Count":                  "50",
			"(Get-NetUDPEndpoint).Count":                    "20",
			"Get-NetTCPConnection -State Listen":            "80\r\n443",
			"(Get-Process).Count":                           "120",
			"Get-Process | Sort-Object":                     "chrome|1234|C:\\chrome.exe|500000|10.5",
			"Win32_BaseBoard":                               "Asus ROG",
			"Win32_BIOS":                                    "AMI|1.0",
			"Win32_ComputerSystemProduct":                   "Asus|Desktop",
			"Get-PhysicalDisk":                              "Samsung SSD 970",
		},
	}

	svc := &Service{}
	status := &HostStatus{}

	res, err := svc.checkWindowsRemote(mockExec, status)
	if err != nil {
		t.Fatalf("checkWindowsRemote failed: %v", err)
	}
	if res.System.Hostname != "win-test-host" {
		t.Errorf("expected hostname win-test-host, got %s", res.System.Hostname)
	}
	if res.System.CurrentUser != "Administrator" {
		t.Errorf("expected user Administrator, got %s", res.System.CurrentUser)
	}
	if res.CPU.Model != "Intel Core i7" {
		t.Errorf("expected cpu model Intel Core i7, got %s", res.CPU.Model)
	}
	if res.CPU.Frequency != 3200 {
		t.Errorf("expected cpu freq 3200, got %f", res.CPU.Frequency)
	}
	if res.Disk.ReadBytes != 1000 {
		t.Errorf("expected disk read bytes 1000, got %d", res.Disk.ReadBytes)
	}
	if res.Network.TotalRx != 10000 {
		t.Errorf("expected net rx 10000, got %d", res.Network.TotalRx)
	}
	if len(res.Network.Interfaces) == 0 || res.Network.Interfaces[0].IP != "192.168.1.10" {
		t.Errorf("expected net interface ip 192.168.1.10, got %v", res.Network.Interfaces)
	}
	if res.Network.TCPConnections != 50 {
		t.Errorf("expected tcp conns 50, got %d", res.Network.TCPConnections)
	}
	if len(res.Network.ListenPorts) != 2 {
		t.Errorf("expected 2 listen ports, got %d", len(res.Network.ListenPorts))
	}
	if res.Process.Total != 120 {
		t.Errorf("expected 120 processes, got %d", res.Process.Total)
	}
	if len(res.Process.List) > 0 && res.Process.List[0].Path != "C:\\chrome.exe" {
		t.Errorf("expected process path C:\\chrome.exe, got %s", res.Process.List[0].Path)
	}
	if res.Hardware.Motherboard != "Asus ROG" {
		t.Errorf("expected mobo Asus ROG, got %s", res.Hardware.Motherboard)
	}
}

func TestCheckLinuxRemote(t *testing.T) {
	mockExec := &MockExecutor{
		Outputs: map[string]string{
			"hostname":   "linux-test-host",
			"uname -sr":  "Linux 5.4.0",
			"uname -m":   "x86_64",
			"uptime -p":  "up 1 hour",
			"cat /proc/uptime": "3600",
			"whoami":     "root",
			"model name": "AMD Ryzen 5",
			"cpu MHz":    "3600.00",
			"nproc":      "4",
			"top -bn1":   "Cpu(s): 10.5 us, 5.0 sy",
			"cat /proc/meminfo": "MemTotal: 16000000 kB\nMemAvailable: 8000000 kB\nSwapTotal: 0 kB\nSwapFree: 0 kB",
			"df -PT -B1":     "ext4 100000000000 50000000000 50000000000 /",
			"cat /proc/diskstats": "8 0 sda 100 0 200 0 50 0 100 0 0 0 0",
			"cat /proc/net/dev": "eth0: 10000 10 0 0 0 0 0 0 5000 5 0 0 0 0 0 0",
			"ip -4 addr show eth0 | grep -oP": "192.168.1.20",
			"ss -s": "Total: 100 (kernel 0)\nTCP: 60 (estab 1, closed 0, orphaned 0, synrecv 0, timewait 0/0), ports 0",
			"netstat -tuln | grep LISTEN": "80",
			"ps -e | wc -l": "150",
			"ps -eo pid,ppid,comm,args,%cpu,%mem": "1 0 systemd /sbin/init 0.1 0.5",
			"cat /sys/class/dmi/id/board_vendor": "Gigabyte",
			"cat /sys/class/dmi/id/board_name": "B450",
			"cat /sys/class/dmi/id/bios_version": "F60",
			"lsblk -d -o MODEL": "WDC WD10EZEX",
		},
	}

	svc := &Service{}
	status := &HostStatus{}

	res, err := svc.checkLinuxRemote(mockExec, status)
	if err != nil {
		t.Fatalf("checkLinuxRemote failed: %v", err)
	}
	if res.System.Hostname != "linux-test-host" {
		t.Errorf("expected hostname linux-test-host, got %s", res.System.Hostname)
	}
	if res.System.OS != "Linux 5.4.0" {
		t.Errorf("expected OS Linux 5.4.0, got %s", res.System.OS)
	}
	if res.CPU.Model != "AMD Ryzen 5" {
		t.Errorf("expected cpu model AMD Ryzen 5, got %s", res.CPU.Model)
	}
	if res.CPU.Frequency != 3600 {
		t.Errorf("expected cpu freq 3600, got %f", res.CPU.Frequency)
	}
	if res.Disk.ReadOps != 100 {
		t.Errorf("expected disk read ops 100, got %d", res.Disk.ReadOps)
	}
	if res.Network.TotalRx != 10000 {
		t.Errorf("expected net rx 10000, got %d", res.Network.TotalRx)
	}
	if len(res.Network.Interfaces) > 0 && res.Network.Interfaces[0].IP != "192.168.1.20" {
		t.Errorf("expected ip 192.168.1.20, got %s", res.Network.Interfaces[0].IP)
	}
	if res.Network.TCPConnections != 60 {
		t.Errorf("expected tcp conns 60, got %d", res.Network.TCPConnections)
	}
	if len(res.Network.ListenPorts) != 1 || res.Network.ListenPorts[0] != 80 {
		t.Errorf("expected listen port 80, got %v", res.Network.ListenPorts)
	}
	if res.Process.Total != 150 {
		t.Errorf("expected 150 processes, got %d", res.Process.Total)
	}
	if res.Hardware.Motherboard != "Gigabyte B450" {
		t.Errorf("expected mobo Gigabyte B450, got %s", res.Hardware.Motherboard)
	}
}
