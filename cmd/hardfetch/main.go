package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"hardfetch/internal/cli"
	"hardfetch/internal/display"
	"hardfetch/internal/modules/hardware"
	"hardfetch/internal/modules/network"
	"hardfetch/internal/modules/software"
	"hardfetch/internal/modules/system"
	"hardfetch/internal/modules/user"
)

func main() {
	// Define flags
	versionFlag := flag.Bool("version", false, "Print version information")
	helpFlag := flag.Bool("help", false, "Print help information")
	modulesFlag := flag.String("modules", "", "Comma-separated list of modules to show (system,cpu,memory,disk)")
	allFlag := flag.Bool("all", false, "Show all available modules")
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
	fmt.Println("  memory    - Memory information (total, used, available)")
	fmt.Println("  disk      - Disk information (total, used, free)")
	fmt.Println("  network   - Network information (IP addresses, interfaces)")
	fmt.Println("  software  - Software information (shell, editor, packages)")
	fmt.Println("  user      - User information (username, shell, home directory)")
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

func runHardfetch(modulesStr string, showAll, noColors bool) {
	// Get modules to display
	modules := getModulesToDisplay(modulesStr, showAll)

	// Collect and display information for each module
	for _, module := range modules {
		switch module {
		case "system":
			displaySystemInfo(noColors)
		case "cpu":
			displayCPUInfo(noColors)
		case "memory":
			displayMemoryInfo(noColors)
		case "disk":
			displayDiskInfo(noColors)
		case "network":
			displayNetworkInfo(noColors)
		case "software":
			displaySoftwareInfo(noColors)
		case "user":
			displayUserInfo(noColors)
		default:
			fmt.Printf("Unknown module: %s\n", module)
		}
		fmt.Println()
	}
}

func displayUserInfo(noColors bool) {
	info, err := user.GetUserInfo()
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		return
	}

	fmt.Println("User Information:")
	fmt.Println("-----------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("cyan")
	}

	data := map[string]string{
		"Username": info.Username,
		"Name":     info.Name,
		"User ID":  info.UserID,
		"Group ID": info.GroupID,
		"Home Dir": info.HomeDir,
		"Shell":    info.Shell,
		"Hostname": info.Hostname,
	}

	for label, value := range data {
		if value == "" {
			continue
		}
		if noColors {
			fmt.Printf("%-15s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}

	// Show a few important environment variables
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("------------")

	envVars := []string{"PATH", "HOME", "USER", "SHELL", "EDITOR", "TERM"}
	for _, key := range envVars {
		if value, ok := info.Environment[key]; ok {
			if noColors {
				fmt.Printf("%-15s: %s\n", key, value)
			} else {
				fmt.Println(display.FormatInfoWithColor(key, value, color))
			}
		}
	}
	fmt.Println()
}

func getModulesToDisplay(modulesStr string, showAll bool) []string {
	if showAll {
		return []string{"system", "cpu", "memory", "disk", "network", "software", "user"}
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

func displaySystemInfo(noColors bool) {
	info, err := system.GetSystemInfo()
	if err != nil {
		fmt.Printf("Error getting system info: %v\n", err)
		return
	}

	fmt.Println("System Information:")
	fmt.Println("------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("cyan")
	}

	data := map[string]string{
		"OS":         info.OS,
		"Arch":       info.Arch,
		"Kernel":     info.Kernel,
		"Hostname":   info.Hostname,
		"Uptime":     info.FormatUptime(),
		"CPU Cores":  fmt.Sprintf("%d", info.NumCPU),
		"Go Version": info.GoVersion,
	}

	for label, value := range data {
		if noColors {
			fmt.Printf("%-15s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}
	fmt.Println()
}

func displayNetworkInfo(noColors bool) {
	info, err := network.GetNetworkInfo()
	if err != nil {
		fmt.Printf("Error getting network info: %v\n", err)
		return
	}

	fmt.Println("Network Information:")
	fmt.Println("--------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("blue")
	}

	data := map[string]string{
		"Hostname":   info.Hostname,
		"Local IP":   info.LocalIP,
		"Public IP":  info.PublicIP,
		"Interfaces": info.FormatInterfaces(),
	}

	for label, value := range data {
		if noColors {
			fmt.Printf("%-15s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}
	fmt.Println()
}

func displaySoftwareInfo(noColors bool) {
	info, err := software.GetSoftwareInfo()
	if err != nil {
		fmt.Printf("Error getting software info: %v\n", err)
		return
	}

	fmt.Println("Software Information:")
	fmt.Println("---------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("green")
	}

	data := map[string]string{
		"Shell":            info.Shell,
		"Editor":           info.Editor,
		"Go Version":       info.GoVersion,
		"Processes":        fmt.Sprintf("%d", info.ProcessCount),
		"Package Managers": info.FormatPackageManagers(),
	}

	for label, value := range data {
		if noColors {
			fmt.Printf("%-20s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}
	fmt.Println()
}

func displayCPUInfo(noColors bool) {
	hwInfo, err := hardware.GetHardwareInfo()
	if err != nil || hwInfo.CPU == nil {
		fmt.Printf("Error getting CPU info: %v\n", err)
		return
	}

	cpu := hwInfo.CPU
	fmt.Println("CPU Information:")
	fmt.Println("----------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("yellow")
	}

	data := map[string]string{
		"Model":        cpu.Model,
		"Cores":        fmt.Sprintf("%d", cpu.Cores),
		"Threads":      fmt.Sprintf("%d", cpu.Threads),
		"Frequency":    cpu.Frequency,
		"Architecture": cpu.Architecture,
	}

	for label, value := range data {
		if noColors {
			fmt.Printf("%-15s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}
	fmt.Println()
}

func displayMemoryInfo(noColors bool) {
	hwInfo, err := hardware.GetHardwareInfo()
	if err != nil || hwInfo.Memory == nil {
		fmt.Printf("Error getting memory info: %v\n", err)
		return
	}

	mem := hwInfo.Memory
	fmt.Println("Memory Information:")
	fmt.Println("-------------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("green")
	}

	data := map[string]string{
		"Total":     mem.FormatTotal(),
		"Used":      mem.FormatUsed(),
		"Available": mem.FormatAvailable(),
		"Free":      mem.FormatFree(),
	}

	for label, value := range data {
		if noColors {
			fmt.Printf("%-15s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}
	fmt.Println()
}

func displayDiskInfo(noColors bool) {
	hwInfo, err := hardware.GetHardwareInfo()
	if err != nil || hwInfo.Disk == nil {
		fmt.Printf("Error getting disk info: %v\n", err)
		return
	}

	disk := hwInfo.Disk
	fmt.Println("Disk Information:")
	fmt.Println("-----------------")

	color := ""
	if !noColors {
		color = display.GetColorCode("magenta")
	}

	data := map[string]string{
		"Total": disk.FormatTotal(),
		"Used":  disk.FormatUsed(),
		"Free":  disk.FormatFree(),
	}

	for label, value := range data {
		if noColors {
			fmt.Printf("%-15s: %s\n", label, value)
		} else {
			fmt.Println(display.FormatInfoWithColor(label, value, color))
		}
	}
	fmt.Println()
}
