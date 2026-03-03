package rdp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type RDPConfig struct {
	Host     string
	Port     int
	Username string
	Password string // Note: RDP files don't support cleartext password usually, or it's hashed. We might skip password and let user enter it.
	Width    int
	Height   int
}

func GenerateRDPFile(config RDPConfig, path string) error {
	// RDP file format: key:type:value
	// types: i (integer), s (string), b (binary)
	content := fmt.Sprintf(`full address:s:%s:%d
username:s:%s
screen mode id:i:2
desktopwidth:i:%d
desktopheight:i:%d
session bpp:i:32
compression:i:1
keyboardhook:i:2
audiocapturemode:i:0
videoplaybackmode:i:1
connection type:i:2
networkautodetect:i:1
bandwidthautodetect:i:1
displayconnectionbar:i:1
enableworkspacereconnect:i:0
disable wallpaper:i:0
allow font smoothing:i:1
allow desktop composition:i:1
disable full window drag:i:0
disable menu anims:i:0
disable themes:i:0
disable cursor setting:i:0
bitmapcachepersistenable:i:1
audiomode:i:0
redirectprinters:i:0
redirectcomports:i:0
redirectsmartcards:i:0
redirectclipboard:i:1
redirectposdevices:i:0
redirectdrives:i:1
drivestoredirect:s:*
prompt for credentials:i:1
negotiate security layer:i:1
remoteapplicationmode:i:0
alternate shell:s:
shell working directory:s:
gatewayhostname:s:
gatewayusagemethod:i:4
gatewaycredentialssource:i:4
gatewayprofileusagemethod:i:0
promptcredentialonce:i:1
authentication level:i:2
`, config.Host, config.Port, config.Username, config.Width, config.Height)

	return os.WriteFile(path, []byte(content), 0644)
}

func LaunchRDP(config RDPConfig) error {
	// Create a temporary .rdp file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("got0_%s_%d.rdp", config.Host, config.Port))
	if err := GenerateRDPFile(config, tmpFile); err != nil {
		return fmt.Errorf("failed to generate RDP file: %w", err)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("mstsc", tmpFile)
	case "darwin":
		// Microsoft Remote Desktop is usually handled by 'open' if installed.
		// If not, it might try to open with TextEdit.
		cmd = exec.Command("open", tmpFile)
	case "linux":
		// Try generic open first
		cmd = exec.Command("xdg-open", tmpFile)
	default:
		return fmt.Errorf("unsupported OS for RDP launch")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch RDP client: %w", err)
	}

	return nil
}
