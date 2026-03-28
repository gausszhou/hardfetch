package cpuinfo

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Info struct {
	Model        string
	Cores        int
	Threads      int
	Frequency    string
	Architecture string
}

func Get() (*Info, error) {
	switch runtime.GOOS {
	case "windows":
		return getCPUInfoWindows()
	case "darwin":
		return getCPUInfoDarwin()
	case "linux":
		return getCPUInfoLinux()
	default:
		return &Info{
			Model:        "Unknown",
			Cores:        runtime.NumCPU(),
			Threads:      runtime.NumCPU(),
			Architecture: runtime.GOARCH,
		}, nil
	}
}

func getCPUInfoWindows() (*Info, error) {
	info := &Info{}

	kernel32 := windows.MustLoadDLL("kernel32.dll")
	GetSystemInfoProc := kernel32.MustFindProc("GetSystemInfo")

	type systemInfo struct {
		wProcessorArchitecture      uint16
		wReserved                   uint16
		dwPageSize                  uint32
		lpMinimumApplicationAddress uintptr
		lpMaximumApplicationAddress uintptr
		dwActiveProcessorMask       uintptr
		dwNumberOfProcessors        uint32
		dwProcessorType             uint32
		dwAllocationGranularity     uint16
		wProcessorLevel             uint16
		wProcessorRevision          uint16
	}

	var sysinfo systemInfo
	GetSystemInfoProc.Call(uintptr(unsafe.Pointer(&sysinfo)))

	info.Cores = int(sysinfo.dwNumberOfProcessors)
	info.Threads = info.Cores

	processorName := getProcessorNameWindows(kernel32)
	if processorName != "" {
		info.Model = processorName
	}

	info.Architecture = getArchitectureWindows(kernel32)

	info.Frequency = getCPUFrequencyWindows()

	return info, nil
}

func getCPUFrequencyWindows() string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `$ErrorActionPreference='SilentlyContinue';(Get-CimInstance Win32_Processor|Measure-Object -Property MaxClockSpeed -Maximum).Maximum`)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return ""
	}
	freq, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		return ""
	}
	if freq > 0 {
		return fmt.Sprintf("%.2f GHz", freq/1000)
	}
	return ""
}

func getProcessorNameWindows(kernel32 *windows.DLL) string {
	var hKey windows.Handle
	keyPath, err := windows.UTF16PtrFromString(`HARDWARE\DESCRIPTION\System\CentralProcessor\0`)
	if err != nil {
		return "Unknown"
	}

	err = windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, keyPath, 0, windows.KEY_READ, &hKey)
	if err != nil {
		return "Unknown"
	}
	defer windows.RegCloseKey(hKey)

	var buf [256]uint16
	var bufType uint32
	var size = uint32(len(buf) * 2)

	namePtr, _ := windows.UTF16PtrFromString("ProcessorNameString")
	err = windows.RegQueryValueEx(hKey, namePtr, nil, &bufType, (*byte)(unsafe.Pointer(&buf[0])), &size)
	if err != nil {
		return "Unknown"
	}

	return windows.UTF16ToString(buf[:size/2])
}

func getArchitectureWindows(kernel32 *windows.DLL) string {
	isWow64Proc := kernel32.MustFindProc("IsWow64Process")
	var isWow64Flag bool
	ret, _, _ := isWow64Proc.Call(uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&isWow64Flag)))
	if ret == 0 {
		return "x64"
	}

	if isWow64Flag {
		return "x86"
	}
	return "x64"
}

func getCPUInfoDarwin() (*Info, error) {
	info := &Info{
		Architecture: runtime.GOARCH,
	}

	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err == nil {
		info.Model = strings.TrimSpace(string(output))
	}

	cmd = exec.Command("sysctl", "-n", "hw.ncpu")
	output, _ = cmd.Output()
	if cores, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
		info.Cores = cores
		info.Threads = cores
	}

	cmd = exec.Command("sysctl", "-n", "hw.cpufrequency")
	output, _ = cmd.Output()
	if freq, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
		info.Frequency = strconv.FormatInt(freq/1000000, 10) + " MHz"
	}

	return info, nil
}

func getCPUInfoLinux() (*Info, error) {
	info := &Info{
		Architecture: runtime.GOARCH,
	}

	cmd := exec.Command("cat", "/proc/cpuinfo")
	output, err := cmd.Output()
	if err != nil {
		return info, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "processor") {
			info.Threads++
		}
	}

	info.Cores = runtime.NumCPU()

	cmd = exec.Command("lscpu")
	output, _ = cmd.Output()
	lines = strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CPU MHz") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info.Frequency = strings.TrimSpace(parts[1]) + " MHz"
			}
		}
	}

	return info, nil
}
