package hardware

import (
	"fmt"
)

// CPUInfo represents CPU information
type CPUInfo struct {
	Model        string
	Cores        int
	Threads      int
	Frequency    string
	Architecture string
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Free      uint64
}

// DiskInfo represents disk information
type DiskInfo struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// HardwareInfo represents all hardware information
type HardwareInfo struct {
	CPU    *CPUInfo
	Memory *MemoryInfo
	Disk   *DiskInfo
}

// GetHardwareInfo collects hardware information
func GetHardwareInfo() (*HardwareInfo, error) {
	info := &HardwareInfo{}

	// Get CPU info
	if cpuInfo, err := getCPUInfo(); err == nil {
		info.CPU = cpuInfo
	}

	// Get memory info
	if memInfo, err := getMemoryInfo(); err == nil {
		info.Memory = memInfo
	}

	// Get disk info
	if diskInfo, err := getDiskInfo(); err == nil {
		info.Disk = diskInfo
	}

	return info, nil
}

// FormatMemory formats memory size in human readable format
func (m *MemoryInfo) FormatTotal() string {
	return formatBytes(m.Total)
}

func (m *MemoryInfo) FormatUsed() string {
	return formatBytes(m.Used)
}

func (m *MemoryInfo) FormatAvailable() string {
	return formatBytes(m.Available)
}

func (m *MemoryInfo) FormatFree() string {
	return formatBytes(m.Free)
}

// FormatDisk formats disk size in human readable format
func (d *DiskInfo) FormatTotal() string {
	return formatBytes(d.Total)
}

func (d *DiskInfo) FormatUsed() string {
	return formatBytes(d.Used)
}

func (d *DiskInfo) FormatFree() string {
	return formatBytes(d.Free)
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TiB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GiB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MiB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KiB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
