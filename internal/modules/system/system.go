package system

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// SystemInfo represents system information
type SystemInfo struct {
	OS       string
	Arch     string
	Kernel   string
	Hostname string
	Uptime   time.Duration
}

// GetSystemInfo collects system information
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	if hostname, err := getHostname(); err == nil {
		info.Hostname = hostname
	}

	if kernel, err := getKernelVersion(); err == nil {
		info.Kernel = kernel
	}

	if uptime, err := getUptime(); err == nil {
		info.Uptime = uptime
	}

	return info, nil
}

// FormatUptime formats uptime duration to human readable string
func (s *SystemInfo) FormatUptime() string {
	if s.Uptime == 0 {
		return "Unknown"
	}

	days := int(s.Uptime.Hours() / 24)
	hours := int(s.Uptime.Hours()) % 24
	minutes := int(s.Uptime.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d day%s", days, plural(days)))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hour%s", hours, plural(hours)))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d minute%s", minutes, plural(minutes)))
	}

	return strings.Join(parts, ", ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
