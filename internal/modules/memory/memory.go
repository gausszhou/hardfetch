package memory

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Info struct {
	Total     uint64
	Used      uint64
	Available uint64
	Free      uint64
}

type SwapInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Available uint64
}

func Get() (*Info, *SwapInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getMemoryInfoWindows()
	case "darwin":
		return getMemoryInfoDarwin()
	case "linux":
		return getMemoryInfoLinux()
	default:
		return &Info{Total: 0, Used: 0, Available: 0, Free: 0}, &SwapInfo{}, nil
	}
}

func getMemoryInfoWindows() (*Info, *SwapInfo, error) {
	kernel32 := windows.MustLoadDLL("kernel32.dll")
	GlobalMemoryStatusExProc := kernel32.MustFindProc("GlobalMemoryStatusEx")

	type memStatusEx struct {
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

	var memStatus memStatusEx
	memStatus.dwLength = uint32(unsafe.Sizeof(memStatus))

	ret, _, _ := GlobalMemoryStatusExProc.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		return &Info{Total: 0, Used: 0, Available: 0, Free: 0}, &SwapInfo{}, nil
	}

	info := &Info{
		Total:     memStatus.ullTotalPhys,
		Available: memStatus.ullAvailPhys,
		Free:      memStatus.ullAvailPhys,
		Used:      memStatus.ullTotalPhys - memStatus.ullAvailPhys,
	}

	swap := &SwapInfo{
		Total:     memStatus.ullTotalPageFile,
		Available: memStatus.ullAvailPageFile,
		Free:      memStatus.ullAvailPageFile,
		Used:      memStatus.ullTotalPageFile - memStatus.ullAvailPageFile,
	}

	return info, swap, nil
}

func getMemoryInfoDarwin() (*Info, *SwapInfo, error) {
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return &Info{Total: 0, Used: 0, Available: 0, Free: 0}, &SwapInfo{}, nil
	}

	lines := strings.Split(string(output), "\n")
	stats := make(map[string]uint64)

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			key := strings.TrimSpace(strings.ReplaceAll(parts[0], " ", "_"))
			value := strings.TrimSpace(strings.ReplaceAll(parts[1], ".", ""))
			if val, err := strconv.ParseUint(value, 10, 64); err == nil {
				stats[key] = val * 4096
			}
		}
	}

	cmd = exec.Command("sysctl", "-n", "hw.memsize")
	output, _ = cmd.Output()
	total, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)

	active := stats["Pages_active"]
	inactive := stats["Pages_inactive"]
	wired := stats["Pages_wired"]
	free := stats["Pages_free"]

	info := &Info{
		Total:     total,
		Available: free + inactive,
		Free:      free,
		Used:      active + wired,
	}

	cmd = exec.Command("sysctl", "-n", "vm.swapusage")
	output, _ = cmd.Output()
	swapStr := string(output)
	swap := &SwapInfo{}
	if strings.Contains(swapStr, "total=") {
		parts := strings.Fields(swapStr)
		for _, p := range parts {
			if strings.HasPrefix(p, "total=") {
				swap.Total, _ = parseSize(strings.TrimPrefix(p, "total="))
			} else if strings.HasPrefix(p, "used=") {
				swap.Used, _ = parseSize(strings.TrimPrefix(p, "used="))
			} else if strings.HasPrefix(p, "free=") {
				swap.Free, _ = parseSize(strings.TrimPrefix(p, "free="))
			}
		}
		swap.Available = swap.Free
	}

	return info, swap, nil
}

func getMemoryInfoLinux() (*Info, *SwapInfo, error) {
	cmd := exec.Command("cat", "/proc/meminfo")
	output, err := cmd.Output()
	if err != nil {
		return &Info{Total: 0, Used: 0, Available: 0, Free: 0}, &SwapInfo{}, nil
	}

	lines := strings.Split(string(output), "\n")
	stats := make(map[string]uint64)

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(strings.ReplaceAll(parts[1], " kB", ""))
			if val, err := strconv.ParseUint(value, 10, 64); err == nil {
				stats[key] = val * 1024
			}
		}
	}

	info := &Info{
		Total:     stats["MemTotal"],
		Available: stats["MemAvailable"],
		Free:      stats["MemFree"],
		Used:      stats["MemTotal"] - stats["MemFree"] - stats["MemAvailable"] - stats["Buffers"] - stats["Cached"],
	}

	swap := &SwapInfo{
		Total:     stats["SwapTotal"],
		Used:      stats["SwapFree"],
		Free:      stats["SwapTotal"] - stats["SwapFree"],
		Available: stats["SwapFree"],
	}

	return info, swap, nil
}

func parseSize(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "M", "")
	s = strings.ReplaceAll(s, "G", "")
	s = strings.ReplaceAll(s, "K", "")
	s = strings.TrimSpace(s)

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	if strings.Contains(s, "G") {
		return uint64(val * 1024 * 1024 * 1024), nil
	}
	if strings.Contains(s, "M") {
		return uint64(val * 1024 * 1024), nil
	}
	return uint64(val * 1024), nil
}
