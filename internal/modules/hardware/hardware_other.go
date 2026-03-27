//go:build !windows

package hardware

import (
	"runtime"
)

func getCPUInfo() (*CPUInfo, error) {
	// Generic implementation for non-Windows platforms
	info := &CPUInfo{
		Cores:        runtime.NumCPU(),
		Threads:      runtime.NumCPU(),
		Architecture: runtime.GOARCH,
		Model:        "Unknown CPU",
		Frequency:    "Unknown",
	}
	return info, nil
}

func getMemoryInfo() (*MemoryInfo, error) {
	// Generic implementation that returns placeholder data
	// In a real implementation, you would use system-specific calls
	return &MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024, // 16 GB
		Used:      8 * 1024 * 1024 * 1024,  // 8 GB
		Available: 8 * 1024 * 1024 * 1024,  // 8 GB
		Free:      8 * 1024 * 1024 * 1024,  // 8 GB
	}, nil
}

func getDiskInfo() (*DiskInfo, error) {
	// Generic implementation that returns placeholder data
	// In a real implementation, you would use system-specific calls
	return &DiskInfo{
		Total: 256 * 1024 * 1024 * 1024, // 256 GB
		Used:  128 * 1024 * 1024 * 1024, // 128 GB
		Free:  128 * 1024 * 1024 * 1024, // 128 GB
	}, nil
}
