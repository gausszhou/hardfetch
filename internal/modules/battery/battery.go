package battery

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Info struct {
	Percentage    int
	Status        string
	TimeRemaining int
}

func Get() (*Info, error) {
	switch runtime.GOOS {
	case "windows":
		return getBatteryInfoWindows()
	case "darwin":
		return getBatteryInfoDarwin()
	case "linux":
		return getBatteryInfoLinux()
	default:
		return &Info{
			Percentage: 100,
			Status:     "Unknown",
		}, nil
	}
}

func getBatteryInfoWindows() (*Info, error) {
	info := &Info{
		Percentage: 100,
		Status:     "AC Connected",
	}

	kernel32 := windows.MustLoadDLL("kernel32.dll")
	proc := kernel32.MustFindProc("GetSystemPowerStatus")

	type systemPowerStatus struct {
		ACLineStatus        uint8
		BatteryFlag         uint8
		BatteryLifePercent  uint8
		SystemStatusFlag    uint8
		BatteryLifeTime     uint32
		BatteryFullLifeTime uint32
	}

	var result systemPowerStatus
	_, _, _ = proc.Call(uintptr(unsafe.Pointer(&result)))

	if result.BatteryFlag&128 == 0 {
		info.Percentage = int(result.BatteryLifePercent)
		if info.Percentage == 255 {
			info.Percentage = 100
		}

		switch result.ACLineStatus {
		case 0:
			info.Status = "Discharging"
		case 1:
			info.Status = "AC Connected"
		default:
			info.Status = "Unknown"
		}

		if result.BatteryLifeTime != 0xffffffff {
			info.TimeRemaining = int(result.BatteryLifeTime / 60)
		}
	}

	return info, nil
}

func getBatteryInfoDarwin() (*Info, error) {
	info := &Info{
		Percentage: 100,
		Status:     "Unknown",
	}

	cmd := exec.Command("pmset", "-g", "batt")
	output, err := cmd.Output()
	if err != nil {
		return info, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Battery") || strings.Contains(line, "AC") {
			continue
		}
		if strings.Contains(line, "%") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if strings.Contains(p, "%") {
					pct := strings.ReplaceAll(p, "%", "")
					if val, err := strconv.Atoi(pct); err == nil {
						info.Percentage = val
					}
					if i+1 < len(parts) {
						next := parts[i+1]
						if strings.Contains(next, "charging") {
							info.Status = "Charging"
						} else if strings.Contains(next, "discharging") {
							info.Status = "Discharging"
						} else if strings.Contains(next, "charged") || strings.Contains(next, "full") {
							info.Status = "Fully Charged"
						}
					}
					break
				}
			}
		}
	}

	return info, nil
}

func getBatteryInfoLinux() (*Info, error) {
	info := &Info{
		Percentage: 100,
		Status:     "Unknown",
	}

	cmd := exec.Command("cat", "/sys/class/power_supply/BAT0/capacity")
	output, err := cmd.Output()
	if err == nil {
		if val, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			info.Percentage = val
		}
	}

	cmd = exec.Command("cat", "/sys/class/power_supply/BAT0/status")
	output, _ = cmd.Output()
	status := strings.TrimSpace(string(output))
	if strings.Contains(status, "Charging") {
		info.Status = "Charging"
	} else if strings.Contains(status, "Discharging") {
		info.Status = "Discharging"
	} else if strings.Contains(status, "Full") || strings.Contains(status, "charged") {
		info.Status = "Fully Charged"
	} else if strings.Contains(status, "AC") {
		info.Status = "AC Connected"
	}

	return info, nil
}
