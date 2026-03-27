//go:build !windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func getHostname() (string, error) {
	return os.Hostname()
}

func getKernelVersion() (string, error) {
	// Try to get kernel version from uname
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to runtime info
		return runtime.GOOS, nil
	}
	return strings.TrimSpace(string(output)), nil
}

func getUptime() (time.Duration, error) {
	// Try to read uptime from /proc/uptime on Linux
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/uptime")
		if err == nil {
			fields := strings.Fields(string(data))
			if len(fields) > 0 {
				var uptimeSeconds float64
				_, err := fmt.Sscanf(fields[0], "%f", &uptimeSeconds)
				if err == nil {
					return time.Duration(uptimeSeconds * float64(time.Second)), nil
				}
			}
		}
	}

	// Try uptime command as fallback
	cmd := exec.Command("uptime", "-p")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse uptime output like "up 2 days, 3 hours, 30 minutes"
	uptimeStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(uptimeStr, "up ") {
		uptimeStr = uptimeStr[3:]
	}

	// Simple parsing - for production use a more robust parser
	// This is a simplified implementation
	return parseUptimeString(uptimeStr), nil
}

func parseUptimeString(uptimeStr string) time.Duration {
	// Simplified parsing - just return 0 for now
	// In a real implementation, parse days, hours, minutes
	return 0
}
