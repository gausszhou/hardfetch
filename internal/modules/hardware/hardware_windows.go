//go:build windows

package hardware

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW  = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type systemInfo struct {
	wProcessorArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress uintptr
	lpMaximumApplicationAddress uintptr
	dwActiveProcessorMask       uintptr
	dwNumberOfProcessors        uint32
	dwProcessorType             uint32
	dwAllocationGranularity     uint32
	wProcessorLevel             uint16
	wProcessorRevision          uint16
}

func getCPUInfo() (*CPUInfo, error) {
	var sysInfo systemInfo
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&sysInfo)))

	info := &CPUInfo{
		Cores:        int(sysInfo.dwNumberOfProcessors),
		Threads:      int(sysInfo.dwNumberOfProcessors),
		Architecture: getArchitecture(sysInfo.wProcessorArchitecture),
	}

	// Try to get CPU model from registry or WMI
	// For now, use a generic name
	info.Model = "Unknown CPU"

	// Try to get frequency from registry
	// This is a simplified version
	info.Frequency = "Unknown"

	return info, nil
}

func getArchitecture(arch uint16) string {
	switch arch {
	case 0:
		return "x86"
	case 5:
		return "ARM"
	case 6:
		return "IA64"
	case 9:
		return "x64"
	case 12:
		return "ARM64"
	default:
		return "Unknown"
	}
}

func getMemoryInfo() (*MemoryInfo, error) {
	var memInfo memoryStatusEx
	memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))

	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}

	return &MemoryInfo{
		Total:     memInfo.ullTotalPhys,
		Available: memInfo.ullAvailPhys,
		Free:      memInfo.ullAvailPhys,
		Used:      memInfo.ullTotalPhys - memInfo.ullAvailPhys,
	}, nil
}

func getDiskInfo() (*DiskInfo, error) {
	// Get disk info for C: drive (simplified)
	var freeBytes, totalBytes, totalFreeBytes uint64

	// Use current directory's drive
	rootPath := "C:\\"

	ret, _, err := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(rootPath))),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("failed to get disk info: %v", err)
	}

	return &DiskInfo{
		Total: totalBytes,
		Free:  freeBytes,
		Used:  totalBytes - freeBytes,
	}, nil
}
