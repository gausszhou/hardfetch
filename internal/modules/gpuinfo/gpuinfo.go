package gpuinfo

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Info struct {
	Name          string
	Vendor        string
	VRAM          uint64
	DriverVersion string
}

func Get() ([]*Info, error) {
	switch runtime.GOOS {
	case "windows":
		return getGPUInfoWindows()
	case "darwin":
		return getGPUInfoDarwin()
	case "linux":
		return getGPUInfoLinux()
	default:
		return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
	}
}

func getGPUInfoWindows() ([]*Info, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$ErrorActionPreference = 'SilentlyContinue'
		$gpus = Get-CimInstance Win32_VideoController | ForEach-Object {
			@{
				Name = $_.Name
				Vendor = $_.AdapterCompatibility
				VRAM = [int64]$_.AdapterRAM
				DriverVersion = $_.DriverVersion
			}
		}
		$gpus | ConvertTo-Json -Compress -Depth 2
	`)
	output, err := cmd.Output()
	if err != nil {
		return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "null" {
		return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
	}

	type gpuData struct {
		Name          string  `json:"Name"`
		Vendor        string  `json:"Vendor"`
		VRAM          float64 `json:"VRAM"`
		DriverVersion string  `json:"DriverVersion"`
	}

	var gpus []gpuData
	if strings.HasPrefix(outputStr, "[") {
		if err := json.Unmarshal([]byte(outputStr), &gpus); err != nil {
			return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
		}
	} else {
		var gpu gpuData
		if err := json.Unmarshal([]byte(outputStr), &gpu); err != nil {
			return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
		}
		gpus = []gpuData{gpu}
	}

	result := make([]*Info, 0, len(gpus))
	for _, g := range gpus {
		info := &Info{
			Name:          g.Name,
			Vendor:        g.Vendor,
			VRAM:          uint64(g.VRAM),
			DriverVersion: g.DriverVersion,
		}
		if info.Vendor == "" {
			info.Vendor = detectVendor(g.Name)
		}
		result = append(result, info)
	}

	return result, nil
}

func getGPUInfoDarwin() ([]*Info, error) {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
	}

	lines := strings.Split(string(output), "\n")
	result := make([]*Info, 0)
	currentGPU := &Info{Vendor: "Apple"}

	for _, line := range lines {
		if strings.Contains(line, "Chipset Model:") {
			currentGPU = &Info{}
			currentGPU.Name = strings.TrimSpace(strings.Split(line, ":")[1])
			currentGPU.Vendor = "Apple"
		} else if strings.Contains(line, "VRAM") && !strings.Contains(line, "Total") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				vram := strings.TrimSpace(parts[1])
				currentGPU.VRAM = parseVRAM(vram)
			}
		} else if strings.TrimSpace(line) == "" && currentGPU.Name != "" {
			result = append(result, currentGPU)
			currentGPU = &Info{Vendor: "Apple"}
		}
	}
	if currentGPU.Name != "" {
		result = append(result, currentGPU)
	}

	if len(result) == 0 {
		result = append(result, &Info{Name: "Unknown", Vendor: "Unknown"})
	}
	return result, nil
}

func getGPUInfoLinux() ([]*Info, error) {
	cmd := exec.Command("lspci", "-v")
	output, err := cmd.Output()
	if err != nil {
		return []*Info{{Name: "Unknown", Vendor: "Unknown"}}, nil
	}

	lines := strings.Split(string(output), "\n")
	result := make([]*Info, 0)

	for _, line := range lines {
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D controller") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info := &Info{
					Name:   strings.TrimSpace(parts[1]),
					Vendor: detectVendor(parts[1]),
				}
				result = append(result, info)
			}
		}
	}

	if len(result) == 0 {
		result = append(result, &Info{Name: "Unknown", Vendor: "Unknown"})
	}
	return result, nil
}

func parseVRAM(vram string) uint64 {
	vram = strings.TrimSpace(vram)
	vram = strings.ReplaceAll(vram, "MB", "")
	vram = strings.ReplaceAll(vram, "GB", "")
	vram = strings.TrimSpace(vram)

	if val, err := strconv.ParseFloat(vram, 64); err == nil {
		if strings.Contains(vram, "GB") {
			return uint64(val * 1024 * 1024 * 1024)
		}
		return uint64(val * 1024 * 1024)
	}
	return 0
}

func detectVendor(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "nvidia") {
		return "NVIDIA"
	}
	if strings.Contains(lower, "amd") || strings.Contains(lower, "radeon") || strings.Contains(lower, "advanced micro devices") {
		return "AMD"
	}
	if strings.Contains(lower, "intel") {
		return "Intel"
	}
	if strings.Contains(lower, "microsoft") || strings.Contains(lower, "basic render") {
		return "Microsoft"
	}
	if strings.Contains(lower, "apple") {
		return "Apple"
	}
	return "Unknown"
}
