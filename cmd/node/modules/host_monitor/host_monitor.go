//go:build feat_host_monitor

package host_monitor

import (
	"runtime"
	"time"

	"GoT0Emergency/cmd/node/core"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type HostMonitorModule struct{}

func (m *HostMonitorModule) Name() string {
	return "host_monitor"
}

type HostData struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryTotal   uint64  `json:"memory_total"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskUsed      uint64  `json:"disk_used"`
	DiskTotal     uint64  `json:"disk_total"`
	DiskPercent   float64 `json:"disk_percent"`
	Load1         float64 `json:"load_1"`
	Load5         float64 `json:"load_5"`
	Load15        float64 `json:"load_15"`
	NetRx         uint64  `json:"net_rx"`
	NetTx         uint64  `json:"net_tx"`
}

func (m *HostMonitorModule) Run() (interface{}, error) {
	data := HostData{}

	// CPU
	// Use a 1-second interval to get a reliable reading.
	percent, err := cpu.Percent(time.Second, false)
	if err == nil && len(percent) > 0 {
		data.CPUUsage = percent[0]
	}

	// Memory
	vMem, err := mem.VirtualMemory()
	if err == nil {
		data.MemoryUsed = vMem.Used
		data.MemoryTotal = vMem.Total
		data.MemoryPercent = vMem.UsedPercent
	}

	// Disk
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	dUsage, err := disk.Usage(path)
	if err == nil {
		data.DiskUsed = dUsage.Used
		data.DiskTotal = dUsage.Total
		data.DiskPercent = dUsage.UsedPercent
	}

	// Load Avg
	l, err := load.Avg()
	if err == nil {
		data.Load1 = l.Load1
		data.Load5 = l.Load5
		data.Load15 = l.Load15
	}

	// Net IO
	n, err := net.IOCounters(false)
	if err == nil && len(n) > 0 {
		data.NetRx = n[0].BytesRecv
		data.NetTx = n[0].BytesSent
	}

	return data, nil
}

func init() {
	core.RegisterModule(&HostMonitorModule{})
}
