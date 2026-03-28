package gpuinfo

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Info struct {
	Name          string
	VRAM          uint64
	VRAMString    string
	Frequency     string
	Type          string
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
		return []*Info{{Name: "Unknown"}}, nil
	}
}

func getGPUInfoWindows() ([]*Info, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `$ErrorActionPreference='SilentlyContinue';Get-CimInstance Win32_VideoController|Select-Object Name,AdapterRAM,DriverVersion,AdapterDACType|ConvertTo-Json -Compress -Depth 2`)
	output, err := cmd.Output()
	if err != nil {
		return []*Info{{Name: "Unknown"}}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "null" {
		return []*Info{{Name: "Unknown"}}, nil
	}

	type gpuData struct {
		Name           string  `json:"Name"`
		AdapterRAM     float64 `json:"AdapterRAM"`
		DriverVersion  string  `json:"DriverVersion"`
		AdapterDACType string  `json:"AdapterDACType"`
	}

	var gpus []gpuData
	if strings.HasPrefix(outputStr, "[") {
		if err := json.Unmarshal([]byte(outputStr), &gpus); err != nil {
			return []*Info{{Name: "Unknown"}}, nil
		}
	} else {
		var gpu gpuData
		if err := json.Unmarshal([]byte(outputStr), &gpu); err != nil {
			return []*Info{{Name: "Unknown"}}, nil
		}
		gpus = []gpuData{gpu}
	}

	nvidiaFreq, nvidiaVRAM := getNvidiaSMI()

	result := make([]*Info, 0, len(gpus))
	for i, g := range gpus {
		vram := uint64(g.AdapterRAM)
		vramStr := formatVRAM(vram)
		freq := ""
		gpuType := "Integrated"

		if strings.Contains(strings.ToLower(g.Name), "nvidia") || strings.Contains(strings.ToLower(g.Name), "amd") || strings.Contains(strings.ToLower(g.Name), "radeon") {
			gpuType = "Discrete"
		}

		if i < len(nvidiaFreq) {
			freq = nvidiaFreq[i]
		}
		if i < len(nvidiaVRAM) && nvidiaVRAM[i] != "" {
			vramStr = nvidiaVRAM[i]
			vram = uint64(float64(vram) * 1024 * 1024)
		}

		info := &Info{
			Name:          g.Name,
			VRAM:          vram,
			VRAMString:    vramStr,
			Frequency:     freq,
			Type:          gpuType,
			DriverVersion: g.DriverVersion,
		}
		result = append(result, info)
	}

	return result, nil
}

func getNvidiaSMI() ([]string, []string) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,clocks.max.sm,memory.total", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	freqs := make([]string, 0, len(lines))
	vrams := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			freqMHz := strings.TrimSpace(strings.Replace(parts[1], " MHz", "", 1))
			memMiB := strings.TrimSpace(strings.Replace(parts[2], " MiB", "", 1))
			if freq, err := strconv.ParseFloat(freqMHz, 64); err == nil {
				freqs = append(freqs, fmt.Sprintf("%.2f GHz", freq/1000))
			}
			if mem, err := strconv.ParseFloat(memMiB, 64); err == nil {
				vrams = append(vrams, fmt.Sprintf("%.2f GiB", mem/1024))
			}
		}
	}
	return freqs, vrams
}

func formatVRAM(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	gb := float64(bytes) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.2f GiB", gb)
}

func getGPUInfoDarwin() ([]*Info, error) {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return []*Info{{Name: "Unknown"}}, nil
	}

	lines := strings.Split(string(output), "\n")
	result := make([]*Info, 0)
	currentGPU := &Info{}

	for _, line := range lines {
		if strings.Contains(line, "Chipset Model:") {
			currentGPU = &Info{}
			currentGPU.Name = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "VRAM") && !strings.Contains(line, "Total") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				vram := strings.TrimSpace(parts[1])
				currentGPU.VRAM = parseVRAM(vram)
			}
		} else if strings.TrimSpace(line) == "" && currentGPU.Name != "" {
			result = append(result, currentGPU)
			currentGPU = &Info{}
		}
	}
	if currentGPU.Name != "" {
		result = append(result, currentGPU)
	}

	if len(result) == 0 {
		result = append(result, &Info{Name: "Unknown"})
	}
	return result, nil
}

func getGPUInfoLinux() ([]*Info, error) {
	cmd := exec.Command("lspci", "-v")
	output, err := cmd.Output()
	if err != nil {
		return []*Info{{Name: "Unknown"}}, nil
	}

	lines := strings.Split(string(output), "\n")
	result := make([]*Info, 0)

	for _, line := range lines {
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D controller") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info := &Info{
					Name: strings.TrimSpace(parts[1]),
				}
				result = append(result, info)
			}
		}
	}

	if len(result) == 0 {
		result = append(result, &Info{Name: "Unknown"})
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
