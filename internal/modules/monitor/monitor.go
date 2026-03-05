package monitor

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"GoT0Emergency/internal/infra/db"
	"GoT0Emergency/internal/infra/executor"
	"GoT0Emergency/internal/infra/session"
	hostMod "GoT0Emergency/internal/modules/host"
	"GoT0Emergency/internal/modules/monitor/pdh"
	"GoT0Emergency/internal/modules/settings"
	"GoT0Emergency/internal/pkg/log"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// --- Struct Definitions ---

type HostStatus struct {
	System   SystemInfo   `json:"system"`
	CPU      CPUInfo      `json:"cpu"`
	Memory   MemoryInfo   `json:"memory"`
	Disk     DiskInfo     `json:"disk"`
	Network  NetworkInfo  `json:"network"`
	Process  ProcessInfo  `json:"process"`
	Hardware HardwareInfo `json:"hardware"`
}

type SystemInfo struct {
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
	Platform    string `json:"platform"`
	KernelArch  string `json:"kernel_arch"`
	BootTime    uint64 `json:"boot_time"`
	Uptime      uint64 `json:"uptime"`
	UptimeStr   string `json:"uptime_str"`
	CurrentUser string `json:"current_user"`
}

type CPUInfo struct {
	Model         string    `json:"model"`
	CoresLogical  int       `json:"cores_logical"`
	CoresPhysical int       `json:"cores_physical"`
	UsageTotal    float64   `json:"usage_total"`
	UsagePerCore  []float64 `json:"usage_per_core"`
	LoadAvg       string    `json:"load_avg"`
	Frequency     float64   `json:"frequency"` // MHz
}

type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Usage     float64 `json:"usage"`
	SwapTotal uint64  `json:"swap_total"`
	SwapUsed  uint64  `json:"swap_used"`
}

type DiskInfo struct {
	Partitions []PartitionInfo `json:"partitions"`
	Total      uint64          `json:"total"`
	Used       uint64          `json:"used"`
	Usage      float64         `json:"usage"`
	ReadBytes  uint64          `json:"read_bytes"`
	WriteBytes uint64          `json:"write_bytes"`
	ReadOps    uint64          `json:"read_ops"`
	WriteOps   uint64          `json:"write_ops"`
}

type PartitionInfo struct {
	Path   string  `json:"path"`
	FSType string  `json:"fstype"`
	Total  uint64  `json:"total"`
	Used   uint64  `json:"used"`
	Usage  float64 `json:"usage"`
}

type NetworkInfo struct {
	Interfaces     []InterfaceInfo `json:"interfaces"`
	TotalRx        uint64          `json:"total_rx"`
	TotalTx        uint64          `json:"total_tx"`
	TCPConnections int             `json:"tcp_connections"`
	UDPConnections int             `json:"udp_connections"`
	ListenPorts    []int           `json:"listen_ports"`
}

type InterfaceInfo struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Rx   uint64 `json:"rx"`
	Tx   uint64 `json:"tx"`
}

type ProcessInfo struct {
	Total int           `json:"total"`
	List  []ProcessItem `json:"list"`
}

type ProcessItem struct {
	Name string  `json:"name"`
	PID  int32   `json:"pid"`
	PPID int32   `json:"ppid"`
	Path string  `json:"path"`
	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
}

type HardwareInfo struct {
	Motherboard string `json:"motherboard"`
	BIOS        string `json:"bios"`
	BaseBoard   string `json:"baseboard"`
	Chassis     string `json:"chassis"`
	MemoryModel string `json:"memory_model"`
	DiskModel   string `json:"disk_model"`
}

// --- Service ---

type Service struct {
	sessionManager *session.SessionManager
	localExecutor  *executor.LocalExecutor
	hostService    *hostMod.Service
	settings       *settings.Service
	stopChan       chan struct{}
	statusCache    map[int64]*HostStatus
	mu             sync.RWMutex
	pdhQuery       *pdh.PDHQuery // PDH query handle for Windows Queue Length
	pdhCPUQuery    *pdh.PDHQuery // PDH query handle for Windows CPU Utility
}

func NewService(sm *session.SessionManager, le *executor.LocalExecutor, hs *hostMod.Service, set *settings.Service) *Service {
	s := &Service{
		sessionManager: sm,
		localExecutor:  le,
		hostService:    hs,
		settings:       set,
		stopChan:       make(chan struct{}),
		statusCache:    make(map[int64]*HostStatus),
	}

	// Initialize PDH query if on Windows
	if runtime.GOOS == "windows" {
		q, err := pdh.NewProcessorQueueLengthQuery()
		if err == nil {
			s.pdhQuery = q
			log.Info("PDH query initialized successfully for Processor Queue Length")
		} else {
			log.Error("Failed to initialize PDH Queue Length query: " + err.Error())
		}
		
		q2, err := pdh.NewProcessorUtilityQuery()
		if err == nil {
			s.pdhCPUQuery = q2
			log.Info("PDH query initialized successfully for Processor Utility")
		} else {
			// Try fallback to Processor Time
			q3, err := pdh.NewPDHQuery("\\Processor(_Total)\\% Processor Time")
			if err == nil {
				s.pdhCPUQuery = q3
				log.Info("PDH query initialized for Processor Time (fallback)")
			} else {
				log.Error("Failed to initialize PDH CPU query: " + err.Error())
			}
		}
	}

	return s
}

func (s *Service) Start() {
	go s.collectionLoop()
	go s.cleanupLoop()
}

func (s *Service) Stop() {
	close(s.stopChan)
}

func (s *Service) collectionLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run immediately once
	s.collectAll()

	for {
		select {
		case <-ticker.C:
			s.collectAll()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run immediately
	s.cleanup()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Service) collectAll() {
	log.Info("Starting scheduled metrics collection")
	hosts, err := s.hostService.GetHosts()
	if err != nil {
		log.Error("Failed to list hosts for monitoring: " + err.Error())
		return
	}

	for _, h := range hosts {
		// Only check remote hosts
		if h.ID == 0 {
			continue // Skip local if it's in the list? Usually local is treated specially or not in DB.
			// Assuming GetHosts returns only remote hosts from DB.
		}

		// Run checks concurrently? Maybe limit concurrency.
		// For now, serial is safer for resource usage, or use a worker pool.
		// Given "Optimize speed", let's use a semaphore.
		go func(hid int64) {
			_, err := s.CheckHost(hid)
			if err != nil {
				// Log is handled inside CheckHost for critical errors? No, CheckHost returns error.
				// We should log it here.
				log.Error("Failed to check host in background: " + err.Error())
			}
		}(h.ID)
	}
}

func (s *Service) cleanup() {
	hours := s.settings.GetRetentionHours()
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	log.Info(fmt.Sprintf("Cleaning up old metrics, retention_hours: %d, cutoff: %v", hours, cutoff))

	query := "DELETE FROM host_metrics WHERE created_at < ?"
	res, err := db.DB.Exec(query, cutoff)
	if err != nil {
		log.Error("Failed to clean up old metrics: " + err.Error())
		return
	}

	count, _ := res.RowsAffected()
	log.Info("Cleaned up old metrics, deleted_rows: " + strconv.FormatInt(count, 10))
}

func (s *Service) CheckHost(hostID int64) (*HostStatus, error) {
	var status *HostStatus
	var err error
	if hostID == 0 {
		status, err = s.checkLocal()
	} else {
		status, err = s.checkRemote(hostID)
	}

	if err != nil {
		return nil, err
	}

	// Update cache
	s.mu.Lock()
	s.statusCache[hostID] = status
	s.mu.Unlock()

	// Save metrics asynchronously
	// Only save for remote hosts (ID > 0) to avoid FK constraint issues for now
	if hostID > 0 {
		go s.saveMetrics(hostID, status)
	}

	return status, nil
}

func (s *Service) GetStatus(hostID int64) (*HostStatus, error) {
	s.mu.RLock()
	if val, ok := s.statusCache[hostID]; ok {
		s.mu.RUnlock()
		return val, nil
	}
	s.mu.RUnlock()
	// Fallback to sync check
	return s.CheckHost(hostID)
}

func (s *Service) saveMetrics(hostID int64, status *HostStatus) {
	query := `
		INSERT INTO host_metrics (
			host_id, cpu_usage, memory_used, memory_total, 
			disk_used, disk_total, net_rx, net_tx
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.DB.Exec(query,
		hostID,
		status.CPU.UsageTotal,
		status.Memory.Used,
		status.Memory.Total,
		status.Disk.Used,
		status.Disk.Total,
		status.Network.TotalRx,
		status.Network.TotalTx,
	)
	if err != nil {
		log.Error("Failed to save metrics: " + err.Error())
	}
}

type MetricPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
}

func (s *Service) GetHostMetrics(hostID int64, durationStr string) ([]MetricPoint, error) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %w", err)
	}

	cutoff := time.Now().Add(-duration)
	query := `
		SELECT created_at, cpu_usage, memory_used, memory_total, disk_used, disk_total
		FROM host_metrics
		WHERE host_id = ? AND created_at >= ?
		ORDER BY created_at ASC
	`

	rows, err := db.DB.Query(query, hostID, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var points []MetricPoint
	for rows.Next() {
		var createdAt time.Time
		var cpuUsage float64
		var memUsed, memTotal, diskUsed, diskTotal sql.NullInt64

		if err := rows.Scan(&createdAt, &cpuUsage, &memUsed, &memTotal, &diskUsed, &diskTotal); err != nil {
			log.Error("Failed to scan metric row: " + err.Error())
			continue
		}

		point := MetricPoint{
			Timestamp: createdAt,
			CPUUsage:  cpuUsage,
		}

		if memTotal.Valid && memTotal.Int64 > 0 {
			point.MemoryUsage = float64(memUsed.Int64) / float64(memTotal.Int64) * 100
		}
		if diskTotal.Valid && diskTotal.Int64 > 0 {
			point.DiskUsage = float64(diskUsed.Int64) / float64(diskTotal.Int64) * 100
		}

		points = append(points, point)
	}

	return points, nil
}

func (s *Service) checkLocal() (*HostStatus, error) {
	status := &HostStatus{}

	// Force clear cache for local checks if we want truly "fresh" data,
	// but CheckHost updates cache anyway.
	// The issue might be CheckHostStatus in Wails calls GetStatus or CheckHost?
	// App.go usually calls service.CheckHost.
	// Let's add a log to verify it's being called.
	// log.Debug("Checking local host status...")

	// 1. System Info
	hInfo, err := host.Info()
	if err == nil {
		status.System.Hostname = hInfo.Hostname
		status.System.OS = hInfo.OS
		status.System.Platform = fmt.Sprintf("%s %s", hInfo.Platform, hInfo.PlatformVersion)
		status.System.KernelArch = hInfo.KernelArch
		status.System.BootTime = hInfo.BootTime
		status.System.Uptime = hInfo.Uptime
		status.System.UptimeStr = (time.Duration(hInfo.Uptime) * time.Second).String()
	} else {
		// Fallback
		status.System.Hostname, _ = os.Hostname()
		status.System.OS = runtime.GOOS
		status.System.KernelArch = runtime.GOARCH
	}
	// Current User (Simple check)
	status.System.CurrentUser = os.Getenv("USER")
	if status.System.CurrentUser == "" {
		status.System.CurrentUser = os.Getenv("USERNAME") // Windows
	}

	// 2. CPU
	cCounts, _ := cpu.Counts(true)
	status.CPU.CoresLogical = cCounts
	cCountsPhy, _ := cpu.Counts(false)
	status.CPU.CoresPhysical = cCountsPhy

	cInfo, err := cpu.Info()
	if err == nil && len(cInfo) > 0 {
		status.CPU.Model = cInfo[0].ModelName
		status.CPU.Frequency = cInfo[0].Mhz
	}

	// Calculate CPU usage
	// Using a single call with interval to get per-core usage, then calculate total from it.
	// This avoids blocking twice and ensures consistent data.
	cPercentCore, err := cpu.Percent(500*time.Millisecond, true)
	if err == nil {
		status.CPU.UsagePerCore = cPercentCore
		
		// Calculate total usage as average of per-core usage
		var totalUsage float64
		for _, p := range cPercentCore {
			totalUsage += p
		}
		if len(cPercentCore) > 0 {
			status.CPU.UsageTotal = totalUsage / float64(len(cPercentCore))
		}
	} else {
		// Fallback
		cPercent, err := cpu.Percent(500*time.Millisecond, false)
		if err == nil && len(cPercent) > 0 {
			status.CPU.UsageTotal = cPercent[0]
		}
	}
	
	// If on Windows and PDH is available, override Total Usage with PDH value (Task Manager style)
	if runtime.GOOS == "windows" && s.pdhCPUQuery != nil {
		val, err := s.pdhCPUQuery.Collect()
		if err == nil {
			// Cap at 100% just in case, although Utility can go > 100% on some systems with Turbo
			// But for UI consistency, maybe cap it? Task Manager caps at 100%.
			if val > 100 {
				val = 100
			}
			status.CPU.UsageTotal = val
		} else {
			log.Error("PDH CPU Collect failed: " + err.Error())
		}
	}

	if runtime.GOOS == "windows" && s.pdhQuery != nil {
		// Use PDH for Processor Queue Length on Windows
		qLen, err := s.pdhQuery.Collect()
		if err == nil {
			status.CPU.LoadAvg = fmt.Sprintf("%.0f (Queue Length)", qLen)
		} else {
			// Try re-initializing if failed?
			// For now just log
			log.Error("PDH Collect failed: " + err.Error())
			status.CPU.LoadAvg = "N/A"
		}
	} else {
		lAvg, err := load.Avg()
		if err == nil {
			status.CPU.LoadAvg = fmt.Sprintf("%.2f, %.2f, %.2f", lAvg.Load1, lAvg.Load5, lAvg.Load15)
		} else {
			status.CPU.LoadAvg = "N/A"
		}
	}

	// 3. Memory
	vMem, err := mem.VirtualMemory()
	if err == nil {
		status.Memory.Total = vMem.Total
		status.Memory.Used = vMem.Used
		status.Memory.Free = vMem.Available // Available is more accurate for "Free" in OS terms
		status.Memory.Usage = vMem.UsedPercent
	}
	sMem, err := mem.SwapMemory()
	if err == nil {
		status.Memory.SwapTotal = sMem.Total
		status.Memory.SwapUsed = sMem.Used
	}

	// 4. Disk
	parts, err := disk.Partitions(false)
	var totalDisk, usedDisk uint64
	if err == nil {
		for _, p := range parts {
			// Filter out some pseudo-filesystems
			if strings.HasPrefix(p.Mountpoint, "/boot") || strings.HasPrefix(p.Mountpoint, "/snap") {
				continue
			}
			usage, uErr := disk.Usage(p.Mountpoint)
			if uErr == nil {
				status.Disk.Partitions = append(status.Disk.Partitions, PartitionInfo{
					Path:   p.Mountpoint,
					FSType: p.Fstype,
					Total:  usage.Total,
					Used:   usage.Used,
					Usage:  usage.UsedPercent,
				})
				totalDisk += usage.Total
				usedDisk += usage.Used
			}
		}
	}
	status.Disk.Total = totalDisk
	status.Disk.Used = usedDisk
	if totalDisk > 0 {
		status.Disk.Usage = float64(usedDisk) / float64(totalDisk) * 100
	}

	dIOCounters, err := disk.IOCounters()
	if err == nil {
		for _, io := range dIOCounters {
			status.Disk.ReadBytes += io.ReadBytes
			status.Disk.WriteBytes += io.WriteBytes
			status.Disk.ReadOps += io.ReadCount
			status.Disk.WriteOps += io.WriteCount
		}
	}

	// 5. Network
	nIOCounters, err := net.IOCounters(false)
	if err == nil && len(nIOCounters) > 0 {
		status.Network.TotalRx = nIOCounters[0].BytesRecv
		status.Network.TotalTx = nIOCounters[0].BytesSent
	}
	nInterfaces, err := net.Interfaces()
	if err == nil {
		for _, ni := range nInterfaces {
			// Skip loopback and down
			isUp := false
			for _, flag := range ni.Flags {
				if flag == "up" {
					isUp = true
					break
				}
			}
			if !isUp {
				continue
			}

			var ipStr string
			for _, addr := range ni.Addrs {
				if !strings.Contains(addr.Addr, ":") { // IPv4 preference
					ipStr = addr.Addr
					break
				}
			}

			status.Network.Interfaces = append(status.Network.Interfaces, InterfaceInfo{
				Name: ni.Name,
				IP:   ipStr,
			})
		}
	}

	conns, err := net.Connections("tcp")
	if err == nil {
		status.Network.TCPConnections = len(conns)
		for _, c := range conns {
			if c.Status == "LISTEN" {
				status.Network.ListenPorts = append(status.Network.ListenPorts, int(c.Laddr.Port))
			}
		}
	}
	connsUDP, err := net.Connections("udp")
	if err == nil {
		status.Network.UDPConnections = len(connsUDP)
	}

	// 6. Processes
	procs, err := process.Processes()
	if err == nil {
		status.Process.Total = len(procs)
		var pItems []ProcessItem
		for _, p := range procs {
			n, _ := p.Name()
			c, _ := p.CPUPercent()
			m, _ := p.MemoryPercent()
			ppid, _ := p.Ppid()
			path, _ := p.Exe()

			pItems = append(pItems, ProcessItem{
				Name: n,
				PID:  p.Pid,
				PPID: ppid,
				Path: path,
				CPU:  c,
				Mem:  float64(m),
			})
		}
		// Sort by Mem desc
		sort.Slice(pItems, func(i, j int) bool {
			return pItems[i].Mem > pItems[j].Mem
		})
		if len(pItems) > 5 {
			status.Process.List = pItems[:5]
		} else {
			status.Process.List = pItems
		}
	}

	// 8. Hardware (ghw)
	// Note: ghw might require CGO or specific permissions.
	base, err := ghw.Baseboard()
	if err == nil {
		status.Hardware.Motherboard = base.Product
		status.Hardware.BaseBoard = base.Vendor
	}
	bios, err := ghw.BIOS()
	if err == nil {
		status.Hardware.BIOS = fmt.Sprintf("%s %s", bios.Vendor, bios.Version)
	}
	chassis, err := ghw.Chassis()
	if err == nil {
		status.Hardware.Chassis = fmt.Sprintf("%s %s", chassis.Vendor, chassis.Type)
	}
	memHW, err := ghw.Memory()
	if err == nil {
		status.Hardware.MemoryModel = fmt.Sprintf("Total Physical: %d MB", memHW.TotalPhysicalBytes/1024/1024)
	}
	block, err := ghw.Block()
	if err == nil {
		var diskModels []string
		for _, disk := range block.Disks {
			// Clean up model string (remove extra parentheses or weird chars if any)
			model := strings.TrimSpace(disk.Model)
			if model != "" {
				diskModels = append(diskModels, model)
			}
		}
		status.Hardware.DiskModel = strings.Join(diskModels, ", ")
	}

	return status, nil
}

func (s *Service) checkRemote(hostID int64) (*HostStatus, error) {
	if !s.sessionManager.IsConnected(hostID) {
		return nil, fmt.Errorf("host not connected")
	}

	exec, err := s.sessionManager.GetExecutor(hostID)
	if err != nil {
		return nil, err
	}

	status := &HostStatus{}

	// Simple OS Detection
	uname, err := exec.Exec("uname -s")
	if err != nil || strings.Contains(strings.ToLower(uname), "windows") {
		// Assume Windows
		return s.checkWindowsRemote(exec, status)
	}

	// Assume Linux/Unix
	return s.checkLinuxRemote(exec, status)
}

func (s *Service) checkWindowsRemote(exec executor.Executor, status *HostStatus) (*HostStatus, error) {
	// 1. System
	out, _ := exec.Exec("hostname")
	status.System.Hostname = strings.TrimSpace(out)
	status.System.OS = "Windows"
	status.System.Platform = "Windows"

	// Current User
	out, _ = exec.Exec("echo %USERNAME%")
	status.System.CurrentUser = strings.TrimSpace(out)

	// 2. CPU
	out, _ = exec.Exec("powershell \"(Get-CimInstance Win32_Processor).Name\"")
	status.CPU.Model = strings.TrimSpace(out)

	out, _ = exec.Exec("powershell \"(Get-CimInstance Win32_Processor).NumberOfCores\"")
	if val, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		status.CPU.CoresPhysical = val
	}
	out, _ = exec.Exec("powershell \"(Get-CimInstance Win32_Processor).NumberOfLogicalProcessors\"")
	if val, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		status.CPU.CoresLogical = val
	}

	out, _ = exec.Exec("powershell \"(Get-CimInstance Win32_Processor).MaxClockSpeed\"")
	if val, err := strconv.ParseFloat(strings.TrimSpace(out), 64); err == nil {
		status.CPU.Frequency = val
	}

	out, _ = exec.Exec("powershell \"Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average | Select-Object -ExpandProperty Average\"")
	if val, err := strconv.ParseFloat(strings.TrimSpace(out), 64); err == nil {
		status.CPU.UsageTotal = val
	}

	// Per Core Usage
	out, _ = exec.Exec("powershell \"Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor | Where-Object Name -ne '_Total' | Sort-Object -Property Name | Select-Object -ExpandProperty PercentProcessorTime\"")
	cpuLines := strings.Split(strings.TrimSpace(out), "\r\n")
	if len(cpuLines) == 0 || cpuLines[0] == "" {
		cpuLines = strings.Split(strings.TrimSpace(out), "\n")
	}
	var usages []float64
	for _, line := range cpuLines {
		if val, err := strconv.ParseFloat(strings.TrimSpace(line), 64); err == nil {
			usages = append(usages, val)
		}
	}
	status.CPU.UsagePerCore = usages

	// 3. Memory
	out, _ = exec.Exec("powershell \"$os = Get-CimInstance Win32_OperatingSystem; Write-Output $os.TotalVisibleMemorySize; Write-Output $os.FreePhysicalMemory\"")
	lines := strings.Split(strings.TrimSpace(out), "\r\n") // Windows uses CRLF
	if len(lines) < 2 {
		lines = strings.Split(strings.TrimSpace(out), "\n")
	}
	if len(lines) >= 2 {
		t, _ := strconv.ParseUint(strings.TrimSpace(lines[0]), 10, 64)
		f, _ := strconv.ParseUint(strings.TrimSpace(lines[1]), 10, 64)
		status.Memory.Total = t * 1024
		status.Memory.Free = f * 1024
		status.Memory.Used = status.Memory.Total - status.Memory.Free
		if status.Memory.Total > 0 {
			status.Memory.Usage = float64(status.Memory.Used) / float64(status.Memory.Total) * 100
		}
	}

	// Uptime
	out, _ = exec.Exec("powershell \"(Get-Date) - (Get-CimInstance Win32_OperatingSystem).LastBootUpTime | Select-Object -ExpandProperty TotalDays\"")
	if days, err := strconv.ParseFloat(strings.TrimSpace(out), 64); err == nil {
		status.System.UptimeStr = fmt.Sprintf("%.1f days", days)
		status.System.Uptime = uint64(days * 24 * 3600)
		status.System.BootTime = uint64(time.Now().Unix()) - status.System.Uptime
	}

	// 4. Disk
	out, _ = exec.Exec("powershell \"Get-PSDrive -PSProvider FileSystem | Select-Object Name,Used,Free | ForEach-Object { $_.Name + ' ' + $_.Used + ' ' + $_.Free }\"")
	lines = strings.Split(strings.TrimSpace(out), "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		lines = strings.Split(strings.TrimSpace(out), "\n")
	}

	var totalDisk, usedDisk uint64
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			used, _ := strconv.ParseUint(parts[1], 10, 64)
			free, _ := strconv.ParseUint(parts[2], 10, 64)
			total := used + free

			if total > 0 {
				usage := float64(used) / float64(total) * 100
				status.Disk.Partitions = append(status.Disk.Partitions, PartitionInfo{
					Path:   parts[0],
					FSType: "NTFS", // Assumption for Windows usually
					Total:  total,
					Used:   used,
					Usage:  usage,
				})
				totalDisk += total
				usedDisk += used
			}
		}
	}
	status.Disk.Total = totalDisk
	status.Disk.Used = usedDisk
	if totalDisk > 0 {
		status.Disk.Usage = float64(usedDisk) / float64(totalDisk) * 100
	}

	// Disk IO (Counters)
	out, _ = exec.Exec("powershell \"Get-CimInstance Win32_PerfRawData_PerfDisk_PhysicalDisk | Where-Object Name -eq '_Total' | Select-Object -Property DiskReadBytesPerSec, DiskWriteBytesPerSec, DiskReadsPerSec, DiskWritesPerSec | ForEach-Object { $_.DiskReadBytesPerSec + ' ' + $_.DiskWriteBytesPerSec + ' ' + $_.DiskReadsPerSec + ' ' + $_.DiskWritesPerSec }\"")
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) >= 4 {
		rb, _ := strconv.ParseUint(parts[0], 10, 64)
		wb, _ := strconv.ParseUint(parts[1], 10, 64)
		ro, _ := strconv.ParseUint(parts[2], 10, 64)
		wo, _ := strconv.ParseUint(parts[3], 10, 64)
		status.Disk.ReadBytes = rb
		status.Disk.WriteBytes = wb
		status.Disk.ReadOps = ro
		status.Disk.WriteOps = wo
	}

	// 5. Network (Interfaces)
	// Get Bytes
	out, _ = exec.Exec("powershell \"Get-NetAdapterStatistics | Select-Object Name, ReceivedBytes, SentBytes | ForEach-Object { $_.Name + '|' + $_.ReceivedBytes + '|' + $_.SentBytes }\"")
	lines = strings.Split(strings.TrimSpace(out), "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		lines = strings.Split(strings.TrimSpace(out), "\n")
	}

	adapterMap := make(map[string]*InterfaceInfo)
	var totalRx, totalTx uint64

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			rx, _ := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
			tx, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)

			adapterMap[name] = &InterfaceInfo{
				Name: name,
				Rx:   rx,
				Tx:   tx,
			}
			totalRx += rx
			totalTx += tx
		}
	}
	status.Network.TotalRx = totalRx
	status.Network.TotalTx = totalTx

	// Get IPs
	out, _ = exec.Exec("powershell \"Get-NetIPAddress -AddressFamily IPv4 | Select-Object InterfaceAlias, IPAddress | ForEach-Object { $_.InterfaceAlias + '|' + $_.IPAddress }\"")
	lines = strings.Split(strings.TrimSpace(out), "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		lines = strings.Split(strings.TrimSpace(out), "\n")
	}
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[0])
			ip := strings.TrimSpace(parts[1])
			if info, ok := adapterMap[name]; ok {
				info.IP = ip
			}
		}
	}

	// Flatten map to slice
	for _, info := range adapterMap {
		status.Network.Interfaces = append(status.Network.Interfaces, *info)
	}

	// Connections
	out, _ = exec.Exec("powershell \"(Get-NetTCPConnection).Count\"")
	if val, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		status.Network.TCPConnections = val
	}
	out, _ = exec.Exec("powershell \"(Get-NetUDPEndpoint).Count\"")
	if val, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		status.Network.UDPConnections = val
	}

	// Listen Ports
	out, _ = exec.Exec("powershell \"Get-NetTCPConnection -State Listen | Select-Object -Unique LocalPort | Select-Object -ExpandProperty LocalPort\"")
	lines = strings.Split(strings.TrimSpace(out), "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		lines = strings.Split(strings.TrimSpace(out), "\n")
	}
	for _, line := range lines {
		if port, err := strconv.Atoi(strings.TrimSpace(line)); err == nil {
			status.Network.ListenPorts = append(status.Network.ListenPorts, port)
		}
	}
	// 6. Processes
	out, _ = exec.Exec("powershell \"(Get-Process).Count\"")
	if val, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		status.Process.Total = val
	}

	out, _ = exec.Exec("powershell \"Get-Process | Sort-Object -Property WorkingSet -Descending | Select-Object -First 5 -Property ProcessName,Id,Path,WorkingSet,CPU | ForEach-Object { $_.ProcessName + '|' + $_.Id + '|' + $_.Path + '|' + $_.WorkingSet + '|' + $_.CPU }\"")
	lines = strings.Split(strings.TrimSpace(out), "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		lines = strings.Split(strings.TrimSpace(out), "\n")
	}

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			name := strings.TrimSpace(parts[0])
			pid, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
			path := strings.TrimSpace(parts[2])
			memBytes, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
			cpuSec, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)

			memPercent := float64(memBytes) / float64(status.Memory.Total) * 100

			status.Process.List = append(status.Process.List, ProcessItem{
				Name: name,
				PID:  int32(pid),
				Path: path,
				CPU:  cpuSec, // Still seconds
				Mem:  memPercent,
			})
		}
	}

	// 7. Hardware
	out, _ = exec.Exec("powershell \"Get-CimInstance Win32_BaseBoard | Select-Object Manufacturer, Product | ForEach-Object { $_.Manufacturer + ' ' + $_.Product }\"")
	status.Hardware.Motherboard = strings.TrimSpace(out)

	out, _ = exec.Exec("powershell \"Get-CimInstance Win32_BIOS | Select-Object Manufacturer, SMBIOSBIOSVersion | ForEach-Object { $_.Manufacturer + ' ' + $_.SMBIOSBIOSVersion }\"")
	status.Hardware.BIOS = strings.TrimSpace(out)

	out, _ = exec.Exec("powershell \"Get-CimInstance Win32_ComputerSystemProduct | Select-Object Name, Vendor | ForEach-Object { $_.Vendor + ' ' + $_.Name }\"")
	status.Hardware.BaseBoard = strings.TrimSpace(out) // Use System Product as BaseBoard/Chassis proxy

	out, _ = exec.Exec("powershell \"Get-PhysicalDisk | Select-Object Model | ForEach-Object { $_.Model }\"")
	lines = strings.Split(strings.TrimSpace(out), "\r\n")
	var diskModels []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			diskModels = append(diskModels, strings.TrimSpace(line))
		}
	}
	status.Hardware.DiskModel = strings.Join(diskModels, ", ")

	status.Hardware.MemoryModel = fmt.Sprintf("Total: %d MB", status.Memory.Total/1024/1024)

	return status, nil
}

func (s *Service) checkLinuxRemote(exec executor.Executor, status *HostStatus) (*HostStatus, error) {
	// 1. System
	out, _ := exec.Exec("hostname")
	status.System.Hostname = strings.TrimSpace(out)

	out, _ = exec.Exec("uname -sr")
	status.System.OS = strings.TrimSpace(out)

	out, _ = exec.Exec("uname -m")
	status.System.KernelArch = strings.TrimSpace(out)

	out, _ = exec.Exec("uptime -p")
	status.System.UptimeStr = strings.TrimSpace(out)

	// Calculate Uptime seconds from /proc/uptime
	out, _ = exec.Exec("cat /proc/uptime | awk '{print $1}'")
	if val, err := strconv.ParseFloat(strings.TrimSpace(out), 64); err == nil {
		status.System.Uptime = uint64(val)
		status.System.BootTime = uint64(time.Now().Unix()) - status.System.Uptime
	}

	out, _ = exec.Exec("whoami")
	status.System.CurrentUser = strings.TrimSpace(out)

	// 2. CPU
	out, _ = exec.Exec("grep -m1 'model name' /proc/cpuinfo | cut -d: -f2")
	status.CPU.Model = strings.TrimSpace(out)

	// CPU Frequency
	out, _ = exec.Exec("grep -m1 'cpu MHz' /proc/cpuinfo | cut -d: -f2")
	if freq, err := strconv.ParseFloat(strings.TrimSpace(out), 64); err == nil {
		status.CPU.Frequency = freq
	}

	out, _ = exec.Exec("nproc")
	if c, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		status.CPU.CoresLogical = c
	}
	// Physical cores estimate
	out, _ = exec.Exec("grep 'physical id' /proc/cpuinfo | sort -u | wc -l")
	if c, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		out2, _ := exec.Exec("grep 'cpu cores' /proc/cpuinfo | head -1 | awk '{print $4}'")
		coresPerSock, _ := strconv.Atoi(strings.TrimSpace(out2))
		if coresPerSock > 0 {
			status.CPU.CoresPhysical = c * coresPerSock
		} else {
			// Fallback if cpu cores not found
			status.CPU.CoresPhysical = status.CPU.CoresLogical / 2
			if status.CPU.CoresPhysical == 0 {
				status.CPU.CoresPhysical = 1
			}
		}
	}

	// Usage Total & Per Core
	// We use a compound command to get two samples from /proc/stat
	// This adds ~0.5s latency but gives accurate results without external tools
	cmd := "cat /proc/stat | grep cpu; sleep 0.5; cat /proc/stat | grep cpu"
	out, _ = exec.Exec(cmd)
	statLines := strings.Split(strings.TrimSpace(out), "\n")

	// Parse /proc/stat: cpu  user nice system idle iowait irq softirq steal guest guest_nice
	type cpuStat struct {
		idle  float64
		total float64
	}
	parseStat := func(line string) (string, cpuStat) {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			return "", cpuStat{}
		}
		name := parts[0]
		var total float64
		var idle float64
		for i, v := range parts[1:] {
			val, _ := strconv.ParseFloat(v, 64)
			total += val
			if i == 3 { // idle is 4th field (index 3 in parts[1:])
				idle = val
			}
		}
		return name, cpuStat{idle: idle, total: total}
	}

	stats1 := make(map[string]cpuStat)
	stats2 := make(map[string]cpuStat)

	mid := len(statLines) / 2
	for i, line := range statLines {
		name, stat := parseStat(line)
		if name == "" {
			continue
		}
		if i < mid {
			stats1[name] = stat
		} else {
			stats2[name] = stat
		}
	}

	// Calculate usage
	if s1, ok := stats1["cpu"]; ok {
		if s2, ok := stats2["cpu"]; ok {
			totalDelta := s2.total - s1.total
			idleDelta := s2.idle - s1.idle
			if totalDelta > 0 {
				status.CPU.UsageTotal = 100 * (totalDelta - idleDelta) / totalDelta
			}
		}
	}

	// Per Core
	// Logical cores are cpu0, cpu1, ...
	var usages []float64
	// We need to iterate in order cpu0, cpu1...
	for i := 0; i < status.CPU.CoresLogical; i++ {
		name := fmt.Sprintf("cpu%d", i)
		if s1, ok := stats1[name]; ok {
			if s2, ok := stats2[name]; ok {
				totalDelta := s2.total - s1.total
				idleDelta := s2.idle - s1.idle
				usage := 0.0
				if totalDelta > 0 {
					usage = 100 * (totalDelta - idleDelta) / totalDelta
				}
				usages = append(usages, usage)
			} else {
				usages = append(usages, 0)
			}
		} else {
			usages = append(usages, 0)
		}
	}
	status.CPU.UsagePerCore = usages

	out, _ = exec.Exec("cat /proc/loadavg | awk '{print $1, $2, $3}'")
	status.CPU.LoadAvg = strings.TrimSpace(out)

	// 3. Memory
	out, _ = exec.Exec("cat /proc/meminfo")
	memMap := make(map[string]uint64)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.ParseUint(parts[1], 10, 64)
			memMap[key] = val * 1024 // kB to B
		}
	}

	status.Memory.Total = memMap["MemTotal"]
	status.Memory.Free = memMap["MemAvailable"] // Best metric
	if status.Memory.Free == 0 {
		status.Memory.Free = memMap["MemFree"] + memMap["Buffers"] + memMap["Cached"]
	}
	status.Memory.Used = status.Memory.Total - status.Memory.Free
	if status.Memory.Total > 0 {
		status.Memory.Usage = float64(status.Memory.Used) / float64(status.Memory.Total) * 100
	}
	status.Memory.SwapTotal = memMap["SwapTotal"]
	status.Memory.SwapUsed = memMap["SwapTotal"] - memMap["SwapFree"]

	// 4. Disk
	out, _ = exec.Exec("df -PT -B1 | awk 'NR>1 {print $2,$3,$4,$6,$7}'")
	lines = strings.Split(strings.TrimSpace(out), "\n")
	var totalDisk, usedDisk uint64
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 5 {
			fstype := parts[0]
			mount := parts[4]
			if strings.HasPrefix(fstype, "tmpfs") || strings.HasPrefix(fstype, "devtmpfs") || strings.HasPrefix(fstype, "squashfs") || strings.HasPrefix(fstype, "overlay") {
				continue
			}
			total, _ := strconv.ParseUint(parts[1], 10, 64)
			used, _ := strconv.ParseUint(parts[2], 10, 64)

			if total > 0 {
				usage := float64(used) / float64(total) * 100
				status.Disk.Partitions = append(status.Disk.Partitions, PartitionInfo{
					Path:   mount,
					FSType: fstype,
					Total:  total,
					Used:   used,
					Usage:  usage,
				})
				totalDisk += total
				usedDisk += used
			}
		}
	}
	status.Disk.Total = totalDisk
	status.Disk.Used = usedDisk
	if totalDisk > 0 {
		status.Disk.Usage = float64(usedDisk) / float64(totalDisk) * 100
	}

	// Disk Stats (Counters)
	// /proc/diskstats: major minor name reads merged sectors ms ...
	out, _ = exec.Exec("cat /proc/diskstats")
	lines = strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 14 {
			// devName := parts[2]
			// We aggregate all
			reads, _ := strconv.ParseUint(parts[3], 10, 64)
			sectorsRead, _ := strconv.ParseUint(parts[5], 10, 64)
			writes, _ := strconv.ParseUint(parts[7], 10, 64)
			sectorsWritten, _ := strconv.ParseUint(parts[9], 10, 64)

			status.Disk.ReadOps += reads
			status.Disk.WriteOps += writes
			status.Disk.ReadBytes += sectorsRead * 512 // Sector size usually 512
			status.Disk.WriteBytes += sectorsWritten * 512
		}
	}

	// 5. Network
	out, _ = exec.Exec("cat /proc/net/dev | awk 'NR>2 {print $1,$2,$10}'") // Iface, Rx, Tx
	lines = strings.Split(strings.TrimSpace(out), "\n")
	var totalRx, totalTx uint64
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			name := strings.TrimSuffix(parts[0], ":")
			rx, _ := strconv.ParseUint(parts[1], 10, 64)
			tx, _ := strconv.ParseUint(parts[2], 10, 64)

			// Get IP
			ipOut, _ := exec.Exec(fmt.Sprintf("ip -4 addr show %s | grep -oP '(?<=inet\\s)\\d+(\\.\\d+){3}' | head -1", name))
			if ipOut == "" {
				// Fallback
				ipOut, _ = exec.Exec(fmt.Sprintf("ip addr show %s | grep 'inet ' | awk '{print $2}' | cut -d/ -f1", name))
			}

			status.Network.Interfaces = append(status.Network.Interfaces, InterfaceInfo{
				Name: name,
				Rx:   rx,
				Tx:   tx,
				IP:   strings.TrimSpace(ipOut),
			})
			totalRx += rx
			totalTx += tx
		}
	}
	status.Network.TotalRx = totalRx
	status.Network.TotalTx = totalTx

	// Network Connections
	// ss -s or netstat -an
	// ss -s output:
	// Total: ...
	// TCP:   ... (estab 1, closed 0, orphaned 0, synrecv 0, timewait 0/0), ports 0
	out, _ = exec.Exec("ss -s")
	if strings.Contains(out, "TCP:") {
		// Parse rudimentary: "TCP: 5 (estab 1, ...)"
		tcpLines := strings.Split(out, "\n")
		for _, line := range tcpLines {
			if strings.Contains(line, "TCP:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					// "TCP:" might be parts[0] or part of parts[0] depending on spacing
					// usually "TCP: 5" -> parts[0]="TCP:", parts[1]="5"
					valStr := parts[1]
					if val, err := strconv.Atoi(valStr); err == nil {
						status.Network.TCPConnections = val
					}
				}
				break
			}
		}
	} else {
		// Fallback to netstat
		out, _ = exec.Exec("netstat -an | grep tcp | wc -l")
		val, _ := strconv.Atoi(strings.TrimSpace(out))
		status.Network.TCPConnections = val
	}
	out, _ = exec.Exec("netstat -an | grep udp | wc -l")
	val, _ := strconv.Atoi(strings.TrimSpace(out))
	status.Network.UDPConnections = val

	// Listen Ports
	out, _ = exec.Exec("netstat -tuln | grep LISTEN | awk '{print $4}' | awk -F: '{print $NF}' | sort -u")
	lines = strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if port, err := strconv.Atoi(line); err == nil {
			status.Network.ListenPorts = append(status.Network.ListenPorts, port)
		}
	}

	// 6. Processes
	// ps -eo pid,ppid,cmd,%cpu,%mem --sort=-%mem | head -6
	out, _ = exec.Exec("ps -eo pid,ppid,comm,args,%cpu,%mem --sort=-%mem | head -6 | tail -n +2")
	lines = strings.Split(strings.TrimSpace(out), "\n")

	// Get total count
	outCount, _ := exec.Exec("ps -e | wc -l")
	if c, err := strconv.Atoi(strings.TrimSpace(outCount)); err == nil {
		status.Process.Total = c
	}

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			pid, _ := strconv.ParseInt(parts[0], 10, 32)
			ppid, _ := strconv.ParseInt(parts[1], 10, 32)
			name := parts[2]
			// args can be multiple fields, but for now we just want path/cmd
			// cmd is usually parts[3]...
			// But 'args' in ps output includes command + args. 'comm' is short name.
			// Let's assume the path is the first part of args if it starts with /
			path := parts[3]

			// cpu mem are last 2
			memStr := parts[len(parts)-1]
			cpuStr := parts[len(parts)-2]

			c, _ := strconv.ParseFloat(cpuStr, 64)
			m, _ := strconv.ParseFloat(memStr, 64)

			status.Process.List = append(status.Process.List, ProcessItem{
				Name: name,
				PID:  int32(pid),
				PPID: int32(ppid),
				Path: path,
				CPU:  c,
				Mem:  m,
			})
		}
	}

	// 8. Hardware
	// dmidecode usually requires root.
	// Try lshw or /sys/class/dmi/id
	out, _ = exec.Exec("cat /sys/class/dmi/id/board_vendor")
	vendor := strings.TrimSpace(out)
	out, _ = exec.Exec("cat /sys/class/dmi/id/board_name")
	name := strings.TrimSpace(out)
	status.Hardware.Motherboard = fmt.Sprintf("%s %s", vendor, name)

	out, _ = exec.Exec("cat /sys/class/dmi/id/bios_version")
	status.Hardware.BIOS = strings.TrimSpace(out)

	// Disk Model: lsblk -d -o NAME,MODEL
	out, _ = exec.Exec("lsblk -d -o MODEL | tail -n +2")
	lines = strings.Split(strings.TrimSpace(out), "\n")
	var diskModels []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			diskModels = append(diskModels, strings.TrimSpace(line))
		}
	}
	status.Hardware.DiskModel = strings.Join(diskModels, ", ")

	// Memory Model: dmidecode -t memory | grep "Part Number" (Needs root)
	// lshw -short -C memory (Needs root)
	// Fallback to just size
	status.Hardware.MemoryModel = fmt.Sprintf("Total: %d MB", status.Memory.Total/1024/1024)

	return status, nil
}
