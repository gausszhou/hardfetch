package display

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gausszhou/hardfetch/internal/detect"
)

var moduleList = []string{"system", "cpu", "gpu", "memory", "disk", "network", "battery"}

func PrintResult(result *detect.Result) {
	modules := moduleList

	var buffer bytes.Buffer
	for _, line := range getInfoLines(modules, result.System, result.Hardware, result.Network) {
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	fmt.Print(buffer.String())
}

func getInfoLines(modules []string, sysInfo *detect.SystemInfo, hwInfo *detect.HardwareInfo, netInfo *detect.NetworkInfo) []string {
	var lines []string
	var buf bytes.Buffer

	for _, module := range modules {
		buf.Reset()
		switch module {
		case "system":
			displaySystemInfoToBuffer(&buf, sysInfo, nil)
		case "cpu":
			displayCPUInfoToBuffer(&buf, hwInfo, nil)
		case "gpu":
			displayGPUInfoToBuffer(&buf, hwInfo, nil)
		case "memory":
			displayMemoryInfoToBuffer(&buf, hwInfo, nil)
		case "disk":
			displayDiskInfoToBuffer(&buf, hwInfo, nil)
		case "network":
			displayNetworkInfoToBuffer(&buf, netInfo, nil)
		case "battery":
			displayBatteryInfoToBuffer(&buf, hwInfo, nil)
		}
		if buf.Len() > 0 {
			lines = append(lines, strings.TrimSpace(buf.String()))
		}
	}

	var result []string
	for _, block := range lines {
		for _, line := range strings.Split(block, "\n") {
			if line != "" {
				result = append(result, line)
			}
		}
	}
	return result
}

func displaySystemInfoToBuffer(buffer *bytes.Buffer, info *detect.SystemInfo, err error) {
	if err != nil {
		fmt.Fprintf(buffer, "Error getting system info: %v\n", err)
		return
	}

	color := GetColorCode("cyan")

	fields := []struct {
		label string
		value string
	}{
		{"Hostname", info.Hostname},
		{"OS", info.OS},
		{"HOST", info.Host},
		{"Kernel", info.Kernel},
		{"Uptime", info.FormatUptime()},
		{"WM", info.WM},
		{"WM Theme", info.WMTheme},
		{"Theme", info.Theme},
		{"Font", info.Font},
		{"Cursor", info.Cursor},
		{"Terminal", info.Terminal},
		{"Locale", info.Locale},
	}

	for _, field := range fields {
		if field.value == "" {
			continue
		}
		fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
	}
}

func displayNetworkInfoToBuffer(buffer *bytes.Buffer, info *detect.NetworkInfo, err error) {
	if err != nil {
		fmt.Fprintf(buffer, "Error getting network info: %v\n", err)
		return
	}

	color := GetColorCode("blue")

	fields := []struct {
		label string
		value string
	}{
		{"Interfaces", info.FormatInterfaces()},
	}

	for _, field := range fields {
		fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
	}
}

func displayCPUInfoToBuffer(buffer *bytes.Buffer, hwInfo *detect.HardwareInfo, err error) {
	if err != nil || hwInfo.CPU == nil {
		fmt.Fprintf(buffer, "Error getting CPU info: %v\n", err)
		return
	}

	cpu := hwInfo.CPU

	color := GetColorCode("yellow")

	fields := []struct {
		label string
		value string
	}{
		{"CPU", fmt.Sprintf("%s (%d) @ %s", cpu.Model, cpu.Threads, cpu.Frequency)},
	}

	for _, field := range fields {
		fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
	}
}

func displayGPUInfoToBuffer(buffer *bytes.Buffer, hwInfo *detect.HardwareInfo, err error) {
	if err != nil || hwInfo.GPUs == nil || len(hwInfo.GPUs) == 0 {
		return
	}

	color := GetColorCode("red")

	for _, gpu := range hwInfo.GPUs {
		value := gpu.Name
		if gpu.Frequency != "" {
			value += " @ " + gpu.Frequency
		}
		if gpu.VRAMString != "" {
			value += " (" + gpu.VRAMString + ")"
		}
		if gpu.Type != "" {
			value += " [" + gpu.Type + "]"
		}
		fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, "GPU", "\033[0m", value)
	}
}

func displayMemoryInfoToBuffer(buffer *bytes.Buffer, hwInfo *detect.HardwareInfo, err error) {
	if err != nil || hwInfo.Memory == nil {
		fmt.Fprintf(buffer, "Error getting memory info: %v\n", err)
		return
	}

	mem := hwInfo.Memory

	color := GetColorCode("green")

	memPercent := 0
	if mem.Total > 0 {
		memPercent = int(float64(mem.Used) / float64(mem.Total) * 100)
	}

	fields := []struct {
		label string
		value string
	}{
		{"Memory", fmt.Sprintf("%s / %s (%d%%)", mem.FormatUsed(), mem.FormatTotal(), memPercent)},
	}

	if hwInfo.Swap != nil {
		swapPercent := 0
		if hwInfo.Swap.Total > 0 {
			swapPercent = int(float64(hwInfo.Swap.Used) / float64(hwInfo.Swap.Total) * 100)
		}
		fields = append(fields, struct {
			label string
			value string
		}{"Swap", fmt.Sprintf("%s / %s (%d%%)", hwInfo.Swap.FormatUsed(), hwInfo.Swap.FormatTotal(), swapPercent)})
	}

	for _, field := range fields {
		fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
	}
}

func displayDiskInfoToBuffer(buffer *bytes.Buffer, hwInfo *detect.HardwareInfo, err error) {
	if err != nil || hwInfo.Disks == nil || len(hwInfo.Disks) == 0 {
		fmt.Fprintf(buffer, "Error getting disk info: %v\n", err)
		return
	}

	for _, disk := range hwInfo.Disks {
		percent := 0
		if disk.Total > 0 {
			percent = int(float64(disk.Used) / float64(disk.Total) * 100)
		}
		fs := disk.FileSystem
		if fs == "" {
			fs = "NTFS"
		}
		color := GetColorCode("magenta")
		fmt.Fprintf(buffer, "%s%-12s%s: %s / %s (%d%%) - %s\n", color, fmt.Sprintf("Disk (%s)", disk.Drive), "\033[0m", disk.FormatUsed(), disk.FormatTotal(), percent, fs)
	}
}

func displayBatteryInfoToBuffer(buffer *bytes.Buffer, hwInfo *detect.HardwareInfo, err error) {
	if err != nil || hwInfo.Battery == nil {
		return
	}

	color := GetColorCode("green")

	battery := hwInfo.Battery

	fields := []struct {
		label string
		value string
	}{
		{"Battery", fmt.Sprintf("%s [%s]", battery.FormatPercentage(), battery.Status)},
	}

	for _, field := range fields {
		fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
	}
}

func GetColorCode(color string) string {
	switch color {
	case "red":
		return "\033[31m"
	case "green":
		return "\033[32m"
	case "yellow":
		return "\033[33m"
	case "blue":
		return "\033[34m"
	case "magenta":
		return "\033[35m"
	case "cyan":
		return "\033[36m"
	case "white":
		return "\033[37m"
	case "bold":
		return "\033[1m"
	default:
		return ""
	}
}

func FormatInfoWithColor(label, value string, colorCode string) string {
	const reset = "\033[0m"
	return colorCode + label + reset + ": " + value
}
