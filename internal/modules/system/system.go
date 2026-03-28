package system

import (
	"fmt"
	"os"
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
	Host     string
	Uptime   time.Duration
	Shell    string
	Display  string
	WM       string
	WMTheme  string
	Theme    string
	Icons    string
	Font     string
	Cursor   string
	Terminal string
	Locale   string
}

// GetSystemInfo collects system information
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	uptime, _ := getUptime()
	info.Uptime = uptime

	shell, _ := getShell()
	info.Shell = shell

	sysInfo := getAllSystemInfo()
	info.Hostname = sysInfo.Hostname
	info.Host = sysInfo.Model
	info.OS = sysInfo.OSVersion
	info.Kernel = sysInfo.Kernel
	info.Display = sysInfo.Display
	info.WM = sysInfo.WM
	info.WMTheme = sysInfo.WMTheme
	info.Theme = sysInfo.Theme
	info.Font = sysInfo.Font
	info.Cursor = sysInfo.Cursor
	info.Locale = sysInfo.Locale

	info.Icons = "Recycle Bin"
	termEnv := os.Getenv("TERM_PROGRAM")
	if termEnv != "" {
		info.Terminal = termEnv
	} else {
		info.Terminal = "Windows Terminal"
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
