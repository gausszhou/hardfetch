//go:build windows

package hardware

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")
	setupapi = syscall.NewLazyDLL("setupapi.dll")

	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW  = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGetSystemInfo        = kernel32.NewProc("GetSystemInfo")
	procGetLogicalDrives     = kernel32.NewProc("GetLogicalDrives")
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

func getLogicalDrives() []string {
	var drives []string

	// Get logical drives bitmask
	ret, _, _ := procGetLogicalDrives.Call()
	driveMask := uint32(ret)

	// Iterate through drive letters A-Z
	for i := 0; i < 26; i++ {
		if driveMask&(1<<uint(i)) != 0 {
			driveLetter := fmt.Sprintf("%c:", 'A'+i)
			drives = append(drives, driveLetter)
		}
	}

	return drives
}

func getDiskInfoFallback() ([]*DiskInfo, error) {
	// Fallback to C: drive only
	var freeBytes, totalBytes, totalFreeBytes uint64
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

	return []*DiskInfo{{
		Drive: "C:",
		Total: totalBytes,
		Free:  freeBytes,
		Used:  totalBytes - freeBytes,
	}}, nil
}

// getGPUInfoImpl collects GPU information on Windows
func getGPUInfoImpl() ([]*GPUInfo, error) {
	// Try to get GPU info using Windows API
	return getGPUInfoFromAPI()
}

func getGPUInfoFromAPI() ([]*GPUInfo, error) {
	var gpus []*GPUInfo

	// Try to get GPU info using WMI (Windows Management Instrumentation)
	// This uses PowerShell to query Win32_VideoController
	gpus = getGPUInfoFromWMI()

	// If WMI failed, use fallback
	if len(gpus) == 0 {
		gpus = append(gpus, &GPUInfo{
			Name:          "Windows Display Adapter",
			Vendor:        "Microsoft",
			VRAM:          0,
			DriverVersion: "Windows Driver",
		})
	}

	return gpus, nil
}

func getGPUInfoFromWMI() []*GPUInfo {
	var gpus []*GPUInfo

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_VideoController | Select-Object Name, AdapterCompatibility, AdapterRAM, DriverVersion, @{Name="AdapterRAMGB";Expression={[math]::Round($_.AdapterRAM/1GB, 2)}} | ConvertTo-Json -Compress`)

	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("powershell", "-NoProfile", "-Command",
			`Get-WmiObject Win32_VideoController | Select-Object Name, AdapterCompatibility, AdapterRAM, DriverVersion | ConvertTo-Json -Compress`)
		output, err = cmd.Output()
		if err != nil {
			return gpus
		}
	}

	outputStr := string(output)

	// Simple JSON parsing for the array
	// Look for GPU entries
	lines := strings.Split(outputStr, "\n")
	inGPU := false
	currentGPU := &GPUInfo{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "{" {
			// Start of a GPU object
			inGPU = true
			currentGPU = &GPUInfo{}
		} else if line == "}" || line == "}," {
			// End of a GPU object
			if inGPU && currentGPU.Name != "" {
				// Clean up vendor name
				if currentGPU.Vendor == "" {
					currentGPU.Vendor = detectVendorFromName(currentGPU.Name)
				}
				gpus = append(gpus, currentGPU)
			}
			inGPU = false
		} else if inGPU {
			// Parse GPU properties
			if strings.Contains(line, "\"Name\"") {
				currentGPU.Name = extractJSONValue(line, "Name")
			} else if strings.Contains(line, "\"AdapterCompatibility\"") {
				currentGPU.Vendor = extractJSONValue(line, "AdapterCompatibility")
			} else if strings.Contains(line, "\"AdapterRAM\"") {
				vramStr := extractJSONValue(line, "AdapterRAM")
				if vramStr != "" {
					var vram uint64
					fmt.Sscanf(vramStr, "%d", &vram)
					currentGPU.VRAM = vram
				}
			} else if strings.Contains(line, "\"DriverVersion\"") {
				currentGPU.DriverVersion = extractJSONValue(line, "DriverVersion")
			} else if strings.Contains(line, "\"AdapterRAMGB\"") {
				// If we have the GB value, use it
				vramGBStr := extractJSONValue(line, "AdapterRAMGB")
				if vramGBStr != "" && currentGPU.VRAM == 0 {
					var vramGB float64
					fmt.Sscanf(vramGBStr, "%f", &vramGB)
					currentGPU.VRAM = uint64(vramGB * 1024 * 1024 * 1024) // Convert GB to bytes
				}
			}
		}
	}

	return gpus
}

func extractJSONValue(line, key string) string {
	keyPattern := "\"" + key + "\":"
	if !strings.Contains(line, keyPattern) {
		return ""
	}

	// Find the value after the key
	parts := strings.SplitN(line, keyPattern, 2)
	if len(parts) < 2 {
		return ""
	}

	value := strings.TrimSpace(parts[1])
	// Remove trailing comma if present
	value = strings.TrimSuffix(value, ",")
	// Remove quotes
	value = strings.Trim(value, "\"")

	return value
}

func detectVendorFromName(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "nvidia"):
		return "NVIDIA"
	case strings.Contains(name, "amd") || strings.Contains(name, "radeon"):
		return "AMD"
	case strings.Contains(name, "intel"):
		return "Intel"
	case strings.Contains(name, "microsoft"):
		return "Microsoft"
	default:
		return "Unknown"
	}
}

func getDiskInfo() ([]*DiskInfo, error) {
	disks := []*DiskInfo{}

	driveLetters := getLogicalDrives()

	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		Get-WmiObject Win32_LogicalDisk | Where-Object {$_.DriveType -eq 3} | ForEach-Object {
			@{Drive=$_.DeviceID; FS=$_.FileSystem} | ConvertTo-Json -Compress
		}
	`)
	output, _ := cmd.Output()
	fsMap := make(map[string]string)
	outputStr := string(output)
	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"Drive"`) {
			drive := extractJSONValue(line, "Drive")
			fs := extractJSONValue(line, "FS")
			if drive != "" && fs != "" {
				fsMap[drive] = fs
			}
		}
	}

	for _, drive := range driveLetters {
		var freeBytes, totalBytes, totalFreeBytes uint64
		rootPath := drive + "\\"

		ret, _, _ := procGetDiskFreeSpaceExW.Call(
			uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(rootPath))),
			uintptr(unsafe.Pointer(&freeBytes)),
			uintptr(unsafe.Pointer(&totalBytes)),
			uintptr(unsafe.Pointer(&totalFreeBytes)),
		)

		if ret != 0 && totalBytes > 0 {
			fs := fsMap[drive]
			if fs == "" {
				fs = "NTFS"
			}
			disks = append(disks, &DiskInfo{
				Drive:      drive,
				Total:      totalBytes,
				Free:       freeBytes,
				Used:       totalBytes - freeBytes,
				FileSystem: fs,
			})
		}
	}

	if len(disks) == 0 {
		return getDiskInfoFallback()
	}

	return disks, nil
}

func getSwapInfoImpl() (*SwapInfo, error) {
	var memInfo memoryStatusEx
	memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))

	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get swap info: %v", err)
	}

	return &SwapInfo{
		Total:     memInfo.ullTotalPageFile,
		Free:      memInfo.ullAvailPageFile,
		Used:      memInfo.ullTotalPageFile - memInfo.ullAvailPageFile,
		Available: memInfo.ullAvailPageFile,
	}, nil
}

func getBatteryInfoImpl() (*BatteryInfo, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$battery = Get-CimInstance -ClassName Win32_Battery -ErrorAction SilentlyContinue
		if ($battery) {
			@{
				EstimatedChargeRemaining = $battery.EstimatedChargeRemaining
				BatteryStatus = $battery.BatteryStatus
			} | ConvertTo-Json -Compress
		}
	`)

	output, err := cmd.Output()
	if err != nil {
		return &BatteryInfo{
			Percentage: 100,
			Status:     "AC Connected",
		}, nil
	}

	outputStr := string(output)
	if outputStr == "" || outputStr == "\n" {
		return &BatteryInfo{
			Percentage: 100,
			Status:     "AC Connected",
		}, nil
	}

	var battery struct {
		EstimatedChargeRemaining int `json:"EstimatedChargeRemaining"`
		BatteryStatus            int `json:"BatteryStatus"`
	}

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "EstimatedChargeRemaining") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				pct := strings.TrimSpace(parts[1])
				pct = strings.TrimSuffix(pct, ",")
				fmt.Sscanf(pct, "%d", &battery.EstimatedChargeRemaining)
			}
		}
		if strings.Contains(line, "BatteryStatus") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				status := strings.TrimSpace(parts[1])
				fmt.Sscanf(status, "%d", &battery.BatteryStatus)
			}
		}
	}

	status := getBatteryStatusText(battery.BatteryStatus)

	return &BatteryInfo{
		Percentage: battery.EstimatedChargeRemaining,
		Status:     status,
	}, nil
}

func getBatteryStatusText(status int) string {
	switch status {
	case 1:
		return "Discharging"
	case 2:
		return "AC Connected"
	case 3:
		return "Fully Charged"
	case 4:
		return "Low"
	case 5:
		return "Critical"
	case 6:
		return "Charging"
	case 7:
		return "Charging High"
	case 8:
		return "Charging Low"
	case 9:
		return "Charging Critical"
	case 10:
		return "Undefined"
	case 11:
		return "Partially Charged"
	default:
		return "Unknown"
	}
}
