package battery

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
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

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", `$ErrorActionPreference='SilentlyContinue';Get-CimInstance Win32_Battery|Select-Object EstimatedChargeRemaining,BatteryStatus|ConvertTo-Json -Compress`)
	output, err := cmd.Output()
	if err != nil {
		return info, nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "null" {
		return info, nil
	}

	type batteryData struct {
		EstimatedChargeRemaining int `json:"EstimatedChargeRemaining"`
		BatteryStatus            int `json:"BatteryStatus"`
	}

	var bat batteryData
	if err := json.Unmarshal([]byte(outputStr), &bat); err != nil {
		return info, nil
	}

	info.Percentage = bat.EstimatedChargeRemaining
	switch bat.BatteryStatus {
	case 1:
		info.Status = "Discharging"
	case 2:
		info.Status = "AC Connected"
	case 3:
		info.Status = "Fully Charged"
	case 4:
		info.Status = "Low"
	case 5:
		info.Status = "Critical"
	case 6:
		info.Status = "Charging"
	case 7:
		info.Status = "Charging High"
	case 8:
		info.Status = "Charging Low"
	case 9:
		info.Status = "Charging Critical"
	default:
		info.Status = "Unknown"
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
