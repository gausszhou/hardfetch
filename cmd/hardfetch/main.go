package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gausszhou/hardfetch/internal/cli"
	"github.com/gausszhou/hardfetch/internal/detect"
	"github.com/gausszhou/hardfetch/internal/display"
	"github.com/gausszhou/hardfetch/internal/modules/hardware"
	"github.com/gausszhou/hardfetch/internal/modules/network"
	"github.com/gausszhou/hardfetch/internal/modules/system"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			printVersion()
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	modulesFlag := flag.String("modules", "", "Comma-separated list of modules to show")
	allFlag := flag.Bool("all", true, "Show all available modules")
	noColorsFlag := flag.Bool("no-colors", false, "Disable colors")
	flag.Parse()

	runHardfetch(*modulesFlag, *allFlag, *noColorsFlag)
}

func printVersion() {
	fmt.Printf("%s version %s\n", cli.Name, cli.Version)
	fmt.Printf("Author: %s\n", cli.Author)
	fmt.Printf("Website: %s\n", cli.Website)
}

func printHelp() {
	fmt.Printf("Usage: %s [options]\n", cli.Name)
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hardfetch                    # Show default system information")
	fmt.Println("  hardfetch --modules system,cpu,memory  # Show specific modules")
	fmt.Println("  hardfetch --all              # Show all available modules")
	fmt.Println("  hardfetch --gen-config       # Generate default configuration file")
}

func runHardfetch(modulesStr string, showAll, noColors bool) {
	modules := getModulesToDisplay(modulesStr, showAll)

	var logoLines []string
	var logoWidth int
	if runtime.GOOS == "windows" {
		logoLines = loadLogo("windows")
		logoWidth = getLogoMaxWidth(logoLines)
	}

	result := detect.Detect(detect.GetCoreDetectors()...)

	var buffer bytes.Buffer

	if len(logoLines) > 0 {
		maxLogoHeight := len(logoLines)
		infoLines := getInfoLines(modules, result.System, result.Hardware, result.Network, noColors)

		for i := 0; i < maxLogoHeight || i < len(infoLines); i++ {
			if i < len(logoLines) {
				lineLen := len(logoLines[i])
				buffer.WriteString(logoLines[i])
				if lineLen < logoWidth {
					buffer.WriteString(strings.Repeat(" ", logoWidth-lineLen))
				}
			} else {
				buffer.WriteString(strings.Repeat(" ", logoWidth))
			}

			if i < len(infoLines) {
				buffer.WriteString("  ")
				buffer.WriteString(infoLines[i])
			}
			buffer.WriteString("\n")
		}
	} else {
		for _, line := range getInfoLines(modules, result.System, result.Hardware, result.Network, noColors) {
			buffer.WriteString(line)
			buffer.WriteString("\n")
		}
	}

	fmt.Print(buffer.String())
}

func getLogoMaxWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

func loadLogo(name string) []string {
	logoPath := filepath.Join("logos", name+".txt")
	data, err := os.ReadFile(logoPath)
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}

func getInfoLines(modules []string, sysInfo *system.SystemInfo, hwInfo *hardware.HardwareInfo, netInfo *network.NetworkInfo, noColors bool) []string {
	var lines []string
	var buf bytes.Buffer

	for _, module := range modules {
		buf.Reset()
		switch module {
		case "system":
			displaySystemInfoToBuffer(&buf, sysInfo, nil, noColors)
		case "cpu":
			displayCPUInfoToBuffer(&buf, hwInfo, nil, noColors)
		case "gpu":
			displayGPUInfoToBuffer(&buf, hwInfo, nil, noColors)
		case "memory":
			displayMemoryInfoToBuffer(&buf, hwInfo, nil, noColors)
		case "disk":
			displayDiskInfoToBuffer(&buf, hwInfo, nil, noColors)
		case "network":
			displayNetworkInfoToBuffer(&buf, netInfo, nil, noColors)
		case "battery":
			displayBatteryInfoToBuffer(&buf, hwInfo, nil, noColors)
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

func getModulesToDisplay(modulesStr string, showAll bool) []string {
	if showAll {
		return []string{"system", "cpu", "gpu", "memory", "disk", "network", "battery"}
	}

	if modulesStr == "" {
		return []string{"system", "cpu", "gpu", "memory", "disk", "network", "battery"}
	}

	modules := strings.Split(modulesStr, ",")
	uniqueModules := make(map[string]bool)
	result := []string{}

	for _, module := range modules {
		module = strings.TrimSpace(module)
		if module != "" && !uniqueModules[module] {
			uniqueModules[module] = true
			result = append(result, module)
		}
	}

	return result
}

func displaySystemInfoToBuffer(buffer *bytes.Buffer, info *system.SystemInfo, err error, noColors bool) {
	if err != nil {
		fmt.Fprintf(buffer, "Error getting system info: %v\n", err)
		return
	}

	color := ""
	if !noColors {
		color = display.GetColorCode("cyan")
	}

	fields := []struct {
		label string
		value string
	}{
		{"Hostname", info.Hostname},
		{"OS", info.OS},
		{"Host", info.Host},
		{"Kernel", info.Kernel},
		{"Uptime", info.FormatUptime()},
		{"Shell", info.Shell},
		{"Display", info.Display},
		{"WM", info.WM},
		{"WM Theme", info.WMTheme},
		{"Theme", info.Theme},
		{"Icon", info.Icons},
		{"Font", info.Font},
		{"Cursor", info.Cursor},
		{"Terminal", info.Terminal},
		{"Locale", info.Locale},
	}

	for _, field := range fields {
		if field.value == "" {
			continue
		}
		if noColors {
			fmt.Fprintf(buffer, "%-12s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
}

func displayNetworkInfoToBuffer(buffer *bytes.Buffer, info *network.NetworkInfo, err error, noColors bool) {
	if err != nil {
		fmt.Fprintf(buffer, "Error getting network info: %v\n", err)
		return
	}

	color := ""
	if !noColors {
		color = display.GetColorCode("blue")
	}

	fields := []struct {
		label string
		value string
	}{
		{"Local IP", info.LocalIP},
		{"Public IP", info.PublicIP},
		{"Interfaces", info.FormatInterfaces()},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-12s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
}

func displayCPUInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.CPU == nil {
		fmt.Fprintf(buffer, "Error getting CPU info: %v\n", err)
		return
	}

	cpu := hwInfo.CPU

	color := ""
	if !noColors {
		color = display.GetColorCode("yellow")
	}

	fields := []struct {
		label string
		value string
	}{
		{"CPU", fmt.Sprintf("%s (%d) @ %s", cpu.Model, cpu.Threads, cpu.Frequency)},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-12s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
}

func displayGPUInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.GPUs == nil || len(hwInfo.GPUs) == 0 {
		fmt.Fprintf(buffer, "Error getting GPU info: %v\n", err)
		return
	}

	color := ""
	if !noColors {
		color = display.GetColorCode("yellow")
	}

	for i, gpu := range hwInfo.GPUs {
		if i > 0 {
			fmt.Fprintln(buffer)
		}

		fields := []struct {
			label string
			value string
		}{
			{"GPU", gpu.Name},
		}

		for _, field := range fields {
			if noColors {
				fmt.Fprintf(buffer, "%-12s: %s\n", field.label, field.value)
			} else {
				fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
			}
		}
	}
}

func displayMemoryInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.Memory == nil {
		fmt.Fprintf(buffer, "Error getting memory info: %v\n", err)
		return
	}

	mem := hwInfo.Memory

	color := ""
	if !noColors {
		color = display.GetColorCode("green")
	}

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
		if noColors {
			fmt.Fprintf(buffer, "%-12s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
}

func displayDiskInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
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
		color := ""
		if !noColors {
			color = display.GetColorCode("magenta")
		}
		if noColors {
			fmt.Fprintf(buffer, "%-12s: %s / %s (%d%%) - %s\n", fmt.Sprintf("Disk (%s)", disk.Drive), disk.FormatUsed(), disk.FormatTotal(), percent, fs)
		} else {
			fmt.Fprintf(buffer, "%s%-12s%s: %s / %s (%d%%) - %s\n", color, fmt.Sprintf("Disk (%s)", disk.Drive), "\033[0m", disk.FormatUsed(), disk.FormatTotal(), percent, fs)
		}
	}
}

func displayBatteryInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.Battery == nil {
		return
	}

	color := ""
	if !noColors {
		color = display.GetColorCode("green")
	}

	battery := hwInfo.Battery

	fields := []struct {
		label string
		value string
	}{
		{"Battery", fmt.Sprintf("%s [%s]", battery.FormatPercentage(), battery.Status)},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-12s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-12s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
}
