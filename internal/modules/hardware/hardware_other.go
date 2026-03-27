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

func getDiskInfo() ([]*DiskInfo, error) {
	// Generic implementation that returns placeholder data
	// In a real implementation, you would use system-specific calls
	return []*DiskInfo{{
		Drive: "/",
		Total: 256 * 1024 * 1024 * 1024, // 256 GB
		Used:  128 * 1024 * 1024 * 1024, // 128 GB
		Free:  128 * 1024 * 1024 * 1024, // 128 GB
	}}, nil
}

// getGPUInfoImpl collects GPU information on non-Windows platforms
func getGPUInfoImpl() ([]*GPUInfo, error) {
	// Generic implementation for non-Windows platforms
	// In a real implementation, you would use system-specific calls
	// For Linux: read from /sys/class/drm, lspci, or nvidia-smi
	// For macOS: use IOKit or system_profiler

	// Try to detect GPU based on OS
	var gpuName, vendor string

	switch runtime.GOOS {
	case "linux":
		gpuName = "Linux GPU"
		vendor = "Open Source"
	case "darwin":
		gpuName = "Apple GPU"
		vendor = "Apple"
	default:
		gpuName = "Generic GPU"
		vendor = "Unknown"
	}

	return []*GPUInfo{{
		Name:          gpuName,
		Vendor:        vendor,
		VRAM:          0,
		DriverVersion: "System Driver",
	}}, nil
}
