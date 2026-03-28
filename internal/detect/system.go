package detect

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type SystemInfo struct {
	OS       string
	Arch     string
	Kernel   string
	Hostname string
	Host     string
	Uptime   time.Duration
	Shell    string
	WM       string
	WMTheme  string
	Theme    string
	Font     string
	Cursor   string
	Terminal string
	Locale   string
}

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

func GetSystemInfoDefaults() *SystemInfo {
	return &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}
