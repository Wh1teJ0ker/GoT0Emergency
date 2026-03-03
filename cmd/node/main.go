package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"runtime"
	"time"

	"GoT0Emergency/cmd/node/core"
	_ "GoT0Emergency/cmd/node/modules/host_monitor"
	"GoT0Emergency/internal/pkg/log"
	"GoT0Emergency/internal/pkg/path"

	"github.com/shirou/gopsutil/v3/host"
)

func main() {
	callbackURL := flag.String("callback", "", "Callback URL (e.g. http://localhost:36911/api/callback)")
	interval := flag.Duration("interval", 5*time.Second, "Reporting interval")
	flag.Parse()

	// Initialize Path
	if err := path.Init(); err != nil {
		panic("Failed to init path: " + err.Error())
	}

	// Initialize Log
	if err := log.Init(); err != nil {
		panic("Failed to init log: " + err.Error())
	}

	// If no callback, run once and exit (traditional behavior)
	if *callbackURL == "" {
		runOnce(nil)
		return
	}

	// Loop
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	log.Info("Starting node agent loop", "callback", *callbackURL, "interval", *interval)

	// Initial run
	runOnce(callbackURL)

	for range ticker.C {
		runOnce(callbackURL)
	}
}

func runOnce(callbackURL *string) {
	status := core.AgentStatus{
		OS: runtime.GOOS,
	}

	// Host Info (Always included)
	hInfo, err := host.Info()
	if err == nil {
		status.Hostname = hInfo.Hostname
		status.Platform = hInfo.Platform
		status.Uptime = hInfo.Uptime
	} else {
		status.Hostname, _ = os.Hostname()
	}

	// Collect registered modules
	core.Collect(&status)

	// Output JSON to stdout (optional in loop mode, maybe noisy)
	data, err := json.Marshal(status)
	if err != nil {
		log.Error("Error marshaling json", "err", err)
		return
	}

	if callbackURL == nil {
		os.Stdout.WriteString(string(data) + "\n")
		return
	}

	// Send to callback
	resp, err := http.Post(*callbackURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Error("Failed to send callback", "err", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Error("Callback returned status", "status", resp.Status)
		} else {
			// log.Info("Callback sent successfully") // Verbose
		}
	}
}
