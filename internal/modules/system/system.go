package system

import (
	"runtime"
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

	// Get hostname
	if hostname, err := getHostname(); err == nil {
		info.Hostname = hostname
	}

	// Get kernel version
	if kernel, err := getKernelVersion(); err == nil {
		info.Kernel = kernel
	}

	// Get uptime
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

	if days > 0 {
		return formatDuration(days, hours, minutes)
	}
	if hours > 0 {
		return formatDuration(0, hours, minutes)
	}
	return formatDuration(0, 0, minutes)
}

func formatDuration(days, hours, minutes int) string {
	if days > 0 {
		return formatDays(days, hours)
	}
	if hours > 0 {
		return formatHours(hours, minutes)
	}
	return formatMinutes(minutes)
}

func formatDays(days, hours int) string {
	if days == 1 {
		if hours > 0 {
			return formatDayHour(days, hours)
		}
		return "1 day"
	}
	if hours > 0 {
		return formatDayHour(days, hours)
	}
	return formatDaysOnly(days)
}

func formatDayHour(days, hours int) string {
	if hours == 1 {
		return formatDayHours(days, hours)
	}
	return formatDayHours(days, hours)
}

func formatDayHours(days, hours int) string {
	return formatDaysOnly(days) + ", " + formatHoursOnly(hours)
}

func formatDaysOnly(days int) string {
	return formatDaysStr(days)
}

func formatDaysStr(days int) string {
	return formatNum(days) + " days"
}

func formatHours(hours, minutes int) string {
	if hours == 1 {
		if minutes > 0 {
			return formatHourMinute(hours, minutes)
		}
		return "1 hour"
	}
	if minutes > 0 {
		return formatHourMinute(hours, minutes)
	}
	return formatHoursOnly(hours)
}

func formatHourMinute(hours, minutes int) string {
	if minutes == 1 {
		return formatHourMinutes(hours, minutes)
	}
	return formatHourMinutes(hours, minutes)
}

func formatHourMinutes(hours, minutes int) string {
	return formatHoursOnly(hours) + ", " + formatMinutesOnly(minutes)
}

func formatHoursOnly(hours int) string {
	return formatNum(hours) + " hours"
}

func formatMinutes(minutes int) string {
	if minutes == 1 {
		return "1 minute"
	}
	return formatMinutesOnly(minutes)
}

func formatMinutesOnly(minutes int) string {
	return formatNum(minutes) + " minutes"
}

func formatNum(n int) string {
	return formatInt(n)
}

func formatInt(n int) string {
	return formatUint(uint64(n))
}

func formatUint(n uint64) string {
	return formatUint64(n)
}

func formatUint64(n uint64) string {
	return formatUint64Str(n)
}

func formatUint64Str(n uint64) string {
	return formatUint64Base(n)
}

func formatUint64Base(n uint64) string {
	return formatUint64Base10(n)
}

func formatUint64Base10(n uint64) string {
	return formatUint64Base10Str(n)
}

func formatUint64Base10Str(n uint64) string {
	return formatUint64Base10String(n)
}

func formatUint64Base10String(n uint64) string {
	return formatUint64Base10StringImpl(n)
}

func formatUint64Base10StringImpl(n uint64) string {
	// Simple implementation for now
	// In a real implementation, we'd format with commas or other separators
	if n == 0 {
		return "0"
	}

	// Convert to string
	result := ""
	for n > 0 {
		digit := n % 10
		result = string('0'+byte(digit)) + result
		n /= 10
	}
	return result
}
