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
	Drive      string // Drive letter or mount point (e.g., "C:", "/")
	Total      uint64
	Used       uint64
	Free       uint64
	FileSystem string // File system type (e.g., "NTFS", "FAT32")
}

// GPUInfo represents GPU information
type GPUInfo struct {
	Name          string
	Vendor        string
	VRAM          uint64 // Video memory in bytes
	DriverVersion string
}

// SwapInfo represents swap information
type SwapInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Available uint64
}

// BatteryInfo represents battery information
type BatteryInfo struct {
	Percentage    int
	Status        string
	TimeRemaining int // minutes
}

// HardwareInfo represents all hardware information
type HardwareInfo struct {
	CPU     *CPUInfo
	Memory  *MemoryInfo
	Swap    *SwapInfo
	Disks   []*DiskInfo
	GPUs    []*GPUInfo
	Battery *BatteryInfo
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

	// Get swap info
	if swapInfo, err := getSwapInfo(); err == nil {
		info.Swap = swapInfo
	}

	// Get disk info
	if disks, err := getDiskInfo(); err == nil {
		info.Disks = disks
	}

	// Get GPU info
	if gpus, err := getGPUInfo(); err == nil {
		info.GPUs = gpus
	}

	// Get battery info
	if batteryInfo, err := getBatteryInfo(); err == nil {
		info.Battery = batteryInfo
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

// FormatVRAM formats GPU memory size in human readable format
func (g *GPUInfo) FormatVRAM() string {
	return formatBytes(g.VRAM)
}

// FormatSwap formats swap size in human readable format
func (s *SwapInfo) FormatTotal() string {
	return formatBytes(s.Total)
}

func (s *SwapInfo) FormatUsed() string {
	return formatBytes(s.Used)
}

func (s *SwapInfo) FormatFree() string {
	return formatBytes(s.Free)
}

// FormatBattery formats battery percentage
func (b *BatteryInfo) FormatPercentage() string {
	return fmt.Sprintf("%d%%", b.Percentage)
}

func (b *BatteryInfo) FormatStatus() string {
	return b.Status
}

// getGPUInfo collects GPU information (platform-specific implementation)
func getGPUInfo() ([]*GPUInfo, error) {
	return getGPUInfoImpl()
}

// getSwapInfo collects swap information (platform-specific implementation)
func getSwapInfo() (*SwapInfo, error) {
	return getSwapInfoImpl()
}

// getBatteryInfo collects battery information (platform-specific implementation)
func getBatteryInfo() (*BatteryInfo, error) {
	return getBatteryInfoImpl()
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
