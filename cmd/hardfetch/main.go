package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"hardfetch/internal/cli"
	"hardfetch/internal/display"
	"hardfetch/internal/modules/hardware"
	"hardfetch/internal/modules/network"
	"hardfetch/internal/modules/system"
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
	systemInfo  *system.SystemInfo
	networkInfo *network.NetworkInfo
	hardwareInfo *hardware.HardwareInfo
	systemErr   error
	networkErr  error
	hardwareErr error
}

func runHardfetch(modulesStr string, showAll, noColors bool) {
	modules := getModulesToDisplay(modulesStr, showAll)

	info := &collectedInfo{}
	var wg sync.WaitGroup

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
		default:
			fmt.Fprintf(&buffer, "Unknown module: %s\n", module)
		}
		buffer.WriteString("\n")
	}

	fmt.Print(buffer.String())
}

func getModulesToDisplay(modulesStr string, showAll bool) []string {
	if showAll {
		return []string{"system", "cpu", "gpu", "memory", "disk", "network"}
	}

	if modulesStr == "" {
		// Default modules
		return []string{"system", "cpu", "memory", "disk"}
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

	fmt.Fprintln(buffer, "System Information:")
	fmt.Fprintln(buffer, "------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("cyan")
	}

	fields := []struct {
		label string
		value string
	}{
		{"OS", info.OS},
		{"Arch", info.Arch},
		{"Kernel", info.Kernel},
		{"Hostname", info.Hostname},
		{"Uptime", info.FormatUptime()},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-15s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-15s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
	fmt.Fprintln(buffer)
}


func displayNetworkInfoToBuffer(buffer *bytes.Buffer, info *network.NetworkInfo, err error, noColors bool) {
	if err != nil {
		fmt.Fprintf(buffer, "Error getting network info: %v\n", err)
		return
	}

	fmt.Fprintln(buffer, "Network Information:")
	fmt.Fprintln(buffer, "--------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("blue")
	}

	fields := []struct {
		label string
		value string
	}{
		{"Hostname", info.Hostname},
		{"Local IP", info.LocalIP},
		{"Public IP", info.PublicIP},
		{"Interfaces", info.FormatInterfaces()},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-15s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-15s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
	fmt.Fprintln(buffer)
}


func displayCPUInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.CPU == nil {
		fmt.Fprintf(buffer, "Error getting CPU info: %v\n", err)
		return
	}

	cpu := hwInfo.CPU
	fmt.Fprintln(buffer, "CPU Information:")
	fmt.Fprintln(buffer, "----------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("yellow")
	}

	fields := []struct {
		label string
		value string
	}{
		{"Model", cpu.Model},
		{"Cores", fmt.Sprintf("%d", cpu.Cores)},
		{"Threads", fmt.Sprintf("%d", cpu.Threads)},
		{"Frequency", cpu.Frequency},
		{"Architecture", cpu.Architecture},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-15s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-15s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
	fmt.Fprintln(buffer)
}


func displayGPUInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.GPUs == nil || len(hwInfo.GPUs) == 0 {
		fmt.Fprintf(buffer, "Error getting GPU info: %v\n", err)
		return
	}

	fmt.Fprintln(buffer, "GPU Information:")
	fmt.Fprintln(buffer, "----------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("yellow")
	}

	for i, gpu := range hwInfo.GPUs {
		if i > 0 {
			fmt.Fprintln(buffer)
		}

		gpuLabel := "GPU"
		if len(hwInfo.GPUs) > 1 {
			gpuLabel = fmt.Sprintf("GPU %d", i+1)
		}

		fields := []struct {
			label string
			value string
		}{
			{"Name", gpu.Name},
			{"Vendor", gpu.Vendor},
			{"VRAM", gpu.FormatVRAM()},
			{"Driver", gpu.DriverVersion},
		}

		if noColors {
			fmt.Fprintf(buffer, "%s:\n", gpuLabel)
		} else {
			fmt.Fprintf(buffer, "%s%s%s:\n", color, gpuLabel, "\033[0m")
		}

		for _, field := range fields {
			if noColors {
				fmt.Fprintf(buffer, "  %-15s: %s\n", field.label, field.value)
			} else {
				fmt.Fprintf(buffer, "  %s%-15s%s: %s\n", color, field.label, "\033[0m", field.value)
			}
		}
	}
	fmt.Fprintln(buffer)
}


func displayMemoryInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.Memory == nil {
		fmt.Fprintf(buffer, "Error getting memory info: %v\n", err)
		return
	}

	mem := hwInfo.Memory
	fmt.Fprintln(buffer, "Memory Information:")
	fmt.Fprintln(buffer, "-------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("green")
	}

	fields := []struct {
		label string
		value string
	}{
		{"Total", mem.FormatTotal()},
		{"Used", mem.FormatUsed()},
		{"Available", mem.FormatAvailable()},
		{"Free", mem.FormatFree()},
	}

	for _, field := range fields {
		if noColors {
			fmt.Fprintf(buffer, "%-15s: %s\n", field.label, field.value)
		} else {
			fmt.Fprintf(buffer, "%s%-15s%s: %s\n", color, field.label, "\033[0m", field.value)
		}
	}
	fmt.Fprintln(buffer)
}


func displayDiskInfoToBuffer(buffer *bytes.Buffer, hwInfo *hardware.HardwareInfo, err error, noColors bool) {
	if err != nil || hwInfo.Disks == nil || len(hwInfo.Disks) == 0 {
		fmt.Fprintf(buffer, "Error getting disk info: %v\n", err)
		return
	}

	fmt.Fprintln(buffer, "Disk Information:")
	fmt.Fprintln(buffer, "-----------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("magenta")
	}

	for _, disk := range hwInfo.Disks {
		driveLabel := "Drive"
		if disk.Drive != "" {
			driveLabel = fmt.Sprintf("Drive %s", disk.Drive)
		}

		if noColors {
			fmt.Fprintf(buffer, "%s:\n", driveLabel)
		} else {
			fmt.Fprintf(buffer, "%s%s%s:\n", color, driveLabel, "\033[0m")
		}

		fields := []struct {
			label string
			value string
		}{
			{"Total", disk.FormatTotal()},
			{"Used", disk.FormatUsed()},
			{"Free", disk.FormatFree()},
		}

		for _, field := range fields {
			if noColors {
				fmt.Fprintf(buffer, "  %-15s: %s\n", field.label, field.value)
			} else {
				fmt.Fprintf(buffer, "  %s%-15s%s: %s\n", color, field.label, "\033[0m", field.value)
			}
		}

		fmt.Fprintln(buffer)
	}
}


