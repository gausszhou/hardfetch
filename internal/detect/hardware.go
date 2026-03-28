package detect

import (
	"fmt"
)

type CPUInfo struct {
	Model        string
	Cores        int
	Threads      int
	Frequency    string
	Architecture string
}

type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Free      uint64
}

type DiskInfo struct {
	Drive      string
	Total      uint64
	Used       uint64
	Free       uint64
	FileSystem string
}

type GPUInfo struct {
	Name          string
	Vendor        string
	VRAM          uint64
	DriverVersion string
}

type SwapInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Available uint64
}

type BatteryInfo struct {
	Percentage    int
	Status        string
	TimeRemaining int
}

type HardwareInfo struct {
	CPU     *CPUInfo
	Memory  *MemoryInfo
	Swap    *SwapInfo
	Disks   []*DiskInfo
	GPUs    []*GPUInfo
	Battery *BatteryInfo
}

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

func (d *DiskInfo) FormatTotal() string {
	return formatBytes(d.Total)
}

func (d *DiskInfo) FormatUsed() string {
	return formatBytes(d.Used)
}

func (d *DiskInfo) FormatFree() string {
	return formatBytes(d.Free)
}

func (g *GPUInfo) FormatVRAM() string {
	return formatBytes(g.VRAM)
}

func (s *SwapInfo) FormatTotal() string {
	return formatBytes(s.Total)
}

func (s *SwapInfo) FormatUsed() string {
	return formatBytes(s.Used)
}

func (s *SwapInfo) FormatFree() string {
	return formatBytes(s.Free)
}

func (b *BatteryInfo) FormatPercentage() string {
	return fmt.Sprintf("%d%%", b.Percentage)
}

func (b *BatteryInfo) FormatStatus() string {
	return b.Status
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
