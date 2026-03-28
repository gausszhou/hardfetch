package gpuinfo

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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
	var wg sync.WaitGroup
	var nvidiaGpus, amdGpus, intelGpus []*Info

	wg.Add(3)
	go func() {
		defer wg.Done()
		nvidiaGpus = getNvidiaGPUInfo()
	}()
	go func() {
		defer wg.Done()
		amdGpus = getAmdGPUInfo()
	}()
	go func() {
		defer wg.Done()
		intelGpus = getIntelGPUInfo()
	}()
	wg.Wait()

	result := make([]*Info, 0)
	result = append(result, nvidiaGpus...)
	result = append(result, amdGpus...)
	result = append(result, intelGpus...)

	if len(result) == 0 {
		return []*Info{{Name: "Unknown"}}, nil
	}
	return result, nil
}

func getNvidiaGPUInfo() []*Info {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name,memory.total,clocks.max.sm,driver_version", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	gpus := make([]*Info, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 4 {
			name := strings.TrimSpace(parts[0])
			memStr := strings.TrimSpace(strings.Replace(parts[1], " MiB", "", 1))
			freqStr := strings.TrimSpace(strings.Replace(parts[2], " MHz", "", 1))
			driver := strings.TrimSpace(parts[3])

			vram := uint64(0)
			vramStr := "0 GiB"
			if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
				vram = uint64(mem * 1024 * 1024)
				vramStr = fmt.Sprintf("%.2f GiB", mem/1024)
			}

			freq := ""
			if f, err := strconv.ParseFloat(freqStr, 64); err == nil {
				freq = fmt.Sprintf("%.2f GHz", f/1000)
			}

			gpus = append(gpus, &Info{
				Name:          name,
				VRAM:          vram,
				VRAMString:    vramStr,
				Frequency:     freq,
				Type:          "Discrete",
				DriverVersion: driver,
			})
		}
	}
	return gpus
}

func getAmdGPUInfo() []*Info {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "rocm-smi", "--query-gpu=name,memory.total,clocks.max.memory,driver_version", "--csv")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil
	}

	gpus := make([]*Info, 0, len(lines)-1)
	for i := 1; i < len(lines); i++ {
		parts := strings.Split(lines[i], ",")
		if len(parts) >= 4 {
			name := strings.TrimSpace(parts[0])
			memStr := strings.TrimSpace(strings.Replace(parts[1], " MiB", "", 1))
			freqStr := strings.TrimSpace(strings.Replace(parts[2], " MHz", "", 1))
			driver := strings.TrimSpace(parts[3])

			vram := uint64(0)
			vramStr := "0 GiB"
			if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
				vram = uint64(mem * 1024 * 1024)
				vramStr = fmt.Sprintf("%.2f GiB", mem/1024)
			}

			freq := ""
			if f, err := strconv.ParseFloat(freqStr, 64); err == nil {
				freq = fmt.Sprintf("%.2f GHz", f/1000)
			}

			gpus = append(gpus, &Info{
				Name:          name,
				VRAM:          vram,
				VRAMString:    vramStr,
				Frequency:     freq,
				Type:          "Discrete",
				DriverVersion: driver,
			})
		}
	}
	return gpus
}

func getIntelGPUInfo() []*Info {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		`$ErrorActionPreference='SilentlyContinue';Get-CimInstance Win32_VideoController | Where-Object { $_.Name -like '*Intel*' } | ForEach-Object { "$($_.Name),$($_.AdapterRAM),$($_.DriverVersion)" }`)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil
	}

	lines := strings.Split(outputStr, "\n")
	gpus := make([]*Info, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			memStr := strings.TrimSpace(parts[1])
			driver := strings.TrimSpace(parts[2])

			vram := uint64(0)
			vramStr := "0 GiB"
			if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
				vram = uint64(mem)
				vramStr = fmt.Sprintf("%.2f GiB", mem/1024/1024/1024)
			}

			gpus = append(gpus, &Info{
				Name:          name,
				VRAM:          vram,
				VRAMString:    vramStr,
				Frequency:     "",
				Type:          "Integrated",
				DriverVersion: driver,
			})
		}
	}
	return gpus
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
