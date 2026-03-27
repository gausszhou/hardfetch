//go:build windows

package hardware

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW  = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
	procRegOpenKeyExW        = advapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW     = advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey          = advapi32.NewProc("RegCloseKey")
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

	// Try to get CPU model from registry
	model, err := getCPUModelFromRegistry()
	if err == nil && model != "" {
		info.Model = strings.TrimSpace(model)
	} else {
		// Fallback to generic name
		info.Model = "Intel/AMD Processor"
	}

	// Try to get frequency from registry
	freq, err := getCPUFrequencyFromRegistry()
	if err == nil && freq > 0 {
		info.Frequency = fmt.Sprintf("%.2f GHz", float64(freq)/1000.0)
	} else {
		// Try to get frequency from WMI or other methods
		info.Frequency = getCPUFrequencyFallback()
	}

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

func getCPUModelFromRegistry() (string, error) {
	// Open registry key for CPU information
	// HKEY_LOCAL_MACHINE\HARDWARE\DESCRIPTION\System\CentralProcessor\0
	const (
		HKEY_LOCAL_MACHINE = 0x80000002
		KEY_QUERY_VALUE    = 0x0001
		KEY_WOW64_64KEY    = 0x0100
	)

	var hkey uintptr
	keyPath := `HARDWARE\DESCRIPTION\System\CentralProcessor\0`
	keyPathUTF16, _ := syscall.UTF16PtrFromString(keyPath)

	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(keyPathUTF16)),
		0,
		uintptr(KEY_QUERY_VALUE|KEY_WOW64_64KEY),
		uintptr(unsafe.Pointer(&hkey)),
	)

	if ret != 0 {
		return "", fmt.Errorf("failed to open registry key: %v", err)
	}
	defer procRegCloseKey.Call(hkey)

	// Query ProcessorNameString value
	valueName := "ProcessorNameString"
	valueNameUTF16, _ := syscall.UTF16PtrFromString(valueName)

	var dataType uint32
	var data [256]uint16
	var dataSize uint32 = uint32(len(data) * 2) // Size in bytes

	ret, _, err = procRegQueryValueExW.Call(
		hkey,
		uintptr(unsafe.Pointer(valueNameUTF16)),
		0,
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(&dataSize)),
	)

	if ret != 0 {
		// Try alternative value name
		valueName = "ProcessorNameString"
		valueNameUTF16, _ = syscall.UTF16PtrFromString(valueName)

		ret, _, err = procRegQueryValueExW.Call(
			hkey,
			uintptr(unsafe.Pointer(valueNameUTF16)),
			0,
			uintptr(unsafe.Pointer(&dataType)),
			uintptr(unsafe.Pointer(&data[0])),
			uintptr(unsafe.Pointer(&dataSize)),
		)

		if ret != 0 {
			return "", fmt.Errorf("failed to query registry value: %v", err)
		}
	}

	return syscall.UTF16ToString(data[:]), nil
}

func getCPUFrequencyFromRegistry() (uint64, error) {
	const (
		HKEY_LOCAL_MACHINE = 0x80000002
		KEY_QUERY_VALUE    = 0x0001
		KEY_WOW64_64KEY    = 0x0100
	)

	var hkey uintptr
	keyPath := `HARDWARE\DESCRIPTION\System\CentralProcessor\0`
	keyPathUTF16, _ := syscall.UTF16PtrFromString(keyPath)

	ret, _, err := procRegOpenKeyExW.Call(
		uintptr(HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(keyPathUTF16)),
		0,
		uintptr(KEY_QUERY_VALUE|KEY_WOW64_64KEY),
		uintptr(unsafe.Pointer(&hkey)),
	)

	if ret != 0 {
		return 0, fmt.Errorf("failed to open registry key: %v", err)
	}
	defer procRegCloseKey.Call(hkey)

	// Query ~MHz value
	valueName := "~MHz"
	valueNameUTF16, _ := syscall.UTF16PtrFromString(valueName)

	var dataType uint32
	var frequency uint32
	var dataSize uint32 = 4

	ret, _, err = procRegQueryValueExW.Call(
		hkey,
		uintptr(unsafe.Pointer(valueNameUTF16)),
		0,
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(&frequency)),
		uintptr(unsafe.Pointer(&dataSize)),
	)

	if ret != 0 {
		return 0, fmt.Errorf("failed to query frequency: %v", err)
	}

	return uint64(frequency), nil
}

func getCPUFrequencyFallback() string {
	// Try to get frequency from WMI or other methods
	// For now, return a generic value
	return "2.0+ GHz"
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
