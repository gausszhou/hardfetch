package main

import (
	"bytes"
	"flag"
	"fmt"
	"hardfetch/internal/cli"
	"hardfetch/internal/display"
	"hardfetch/internal/modules/hardware"
	"hardfetch/internal/modules/network"
	"hardfetch/internal/modules/system"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	// Define flags
	versionFlag := flag.Bool("version", false, "Print version information")
	helpFlag := flag.Bool("help", false, "Print help information")
	modulesFlag := flag.String("modules", "", "Comma-separated list of modules to show (system,cpu,memory,disk)")
	allFlag := flag.Bool("all", true, "Show all available modules")
	noColorsFlag := flag.Bool("no-colors", false, "Don't use colors")
	genConfigFlag := flag.Bool("gen-config", false, "Generate default configuration file")
	listModulesFlag := flag.Bool("list-modules", false, "List all available modules")

	flag.Parse()

	// Handle version flag
	if *versionFlag {
		printVersion()
		return
	}

	// Handle help flag
	if *helpFlag {
		printHelp()
		return
	}

	// Handle list modules flag
	if *listModulesFlag {
		printAvailableModules()
		return
	}

	// Handle generate config flag
	if *genConfigFlag {
		generateConfig()
		return
	}

	// Main execution
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

func printAvailableModules() {
	fmt.Println("Available modules:")
	fmt.Println("  system    - System information (OS, kernel, hostname, uptime)")
	fmt.Println("  cpu       - CPU information (model, cores, frequency)")
	fmt.Println("  gpu       - GPU information (name, vendor, VRAM, driver)")
	fmt.Println("  memory    - Memory information (total, used, available)")
	fmt.Println("  disk      - Disk information (total, used, free)")
	fmt.Println("  network   - Network information (IP addresses, interfaces)")
}

func generateConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := fmt.Sprintf("%s/.config/hardfetch/config.json", home)
	if err := cli.GenerateDefaultConfig(configPath); err != nil {
		fmt.Printf("Error generating config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Default configuration generated at: %s\n", configPath)
	fmt.Println("You can edit this file to customize hardfetch behavior.")
}

type collectedInfo struct {
	systemInfo   *system.SystemInfo
	networkInfo  *network.NetworkInfo
	hardwareInfo *hardware.HardwareInfo
	systemErr    error
	networkErr   error
	hardwareErr  error
}

func runHardfetch(modulesStr string, showAll, noColors bool) {
	modules := getModulesToDisplay(modulesStr, showAll)

	info := &collectedInfo{}
	var wg sync.WaitGroup

	stopLoading := make(chan struct{})

	go func() {
		dots := []string{".  ", ".. ", "..."}
		idx := 0
		for {
			select {
			case <-stopLoading:
				return
			default:
				fmt.Fprintf(os.Stderr, "\rCollecting%s", dots[idx%len(dots)])
				idx++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	wg.Add(3)
	go func() {
		defer wg.Done()
		info.systemInfo, info.systemErr = system.GetSystemInfo()
	}()
	go func() {
		defer wg.Done()
		info.networkInfo, info.networkErr = network.GetNetworkInfo()
	}()
	go func() {
		defer wg.Done()
		info.hardwareInfo, info.hardwareErr = hardware.GetHardwareInfo()
	}()

	wg.Wait()
	close(stopLoading)
	fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", 15)+"\r")

	var buffer bytes.Buffer

	for _, module := range modules {
		switch module {
		case "system":
			displaySystemInfoToBuffer(&buffer, info.systemInfo, info.systemErr, noColors)
		case "cpu":
			displayCPUInfoToBuffer(&buffer, info.hardwareInfo, info.hardwareErr, noColors)
		case "gpu":
			displayGPUInfoToBuffer(&buffer, info.hardwareInfo, info.hardwareErr, noColors)
		case "memory":
			displayMemoryInfoToBuffer(&buffer, info.hardwareInfo, info.hardwareErr, noColors)
		case "disk":
			displayDiskInfoToBuffer(&buffer, info.hardwareInfo, info.hardwareErr, noColors)
		case "network":
			displayNetworkInfoToBuffer(&buffer, info.networkInfo, info.networkErr, noColors)
		case "battery":
			displayBatteryInfoToBuffer(&buffer, info.hardwareInfo, info.hardwareErr, noColors)
		default:
			fmt.Fprintf(&buffer, "Unknown module: %s\n", module)
		}
	}

	fmt.Print(buffer.String())
}

func getModulesToDisplay(modulesStr string, showAll bool) []string {
	if showAll {
		return []string{"system", "cpu", "gpu", "memory", "disk", "network", "battery"}
	}

	if modulesStr == "" {
		// Default modules
		return []string{"system", "cpu", "gpu", "memory", "disk", "network", "battery"}
	}

	modules := strings.Split(modulesStr, ",")
	// Remove empty strings and duplicates
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
