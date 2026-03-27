package software

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PackageManager represents a package manager
type PackageManager struct {
	Name    string
	Command string
	Count   int
}

// ProcessInfo represents a running process
type ProcessInfo struct {
	PID    int
	Name   string
	User   string
	CPU    float64
	Memory float64
}

// SoftwareInfo represents software information
type SoftwareInfo struct {
	PackageManagers []PackageManager
	ProcessCount    int
	Shell           string
	Editor          string
	GoVersion       string
}

// GetSoftwareInfo collects software information
func GetSoftwareInfo() (*SoftwareInfo, error) {
	info := &SoftwareInfo{}

	// Get shell
	if shell := getShell(); shell != "" {
		info.Shell = shell
	}

	// Get editor
	if editor := getEditor(); editor != "" {
		info.Editor = editor
	}

	// Get Go version
	info.GoVersion = runtime.Version()

	// Get package managers
	info.PackageManagers = getPackageManagers()

	// Get process count (simplified)
	info.ProcessCount = getProcessCount()

	return info, nil
}

func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Try COMSPEC on Windows
		shell = os.Getenv("COMSPEC")
	}
	return shell
}

func getEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	return editor
}

func getPackageManagers() []PackageManager {
	var managers []PackageManager

	// Check for common package managers
	checks := []struct {
		name    string
		command string
		args    []string
	}{
		{"apt", "apt", []string{"list", "--installed"}},
		{"yum", "yum", []string{"list", "installed"}},
		{"dnf", "dnf", []string{"list", "installed"}},
		{"pacman", "pacman", []string{"-Q"}},
		{"brew", "brew", []string{"list"}},
		{"npm", "npm", []string{"list", "-g", "--depth=0"}},
		{"pip", "pip", []string{"list"}},
		{"cargo", "cargo", []string{"install", "--list"}},
	}

	for _, check := range checks {
		if _, err := exec.LookPath(check.command); err == nil {
			// Try to run the command to check if it works
			cmd := exec.Command(check.command, check.args...)
			output, err := cmd.Output()
			if err == nil {
				// Count lines (rough estimate of packages)
				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				count := len(lines) - 1 // Subtract header line
				if count < 0 {
					count = 0
				}

				managers = append(managers, PackageManager{
					Name:    check.name,
					Command: check.command,
					Count:   count,
				})
			}
		}
	}

	return managers
}

func getProcessCount() int {
	// Simplified implementation
	// On Unix-like systems, we can check /proc
	if runtime.GOOS != "windows" {
		if _, err := os.Stat("/proc"); err == nil {
			entries, err := os.ReadDir("/proc")
			if err == nil {
				count := 0
				for _, entry := range entries {
					// Check if entry name is numeric (PID)
					if _, err := fmt.Sscanf(entry.Name(), "%d", new(int)); err == nil {
						count++
					}
				}
				return count
			}
		}
	}

	// Fallback: use ps command
	cmd := exec.Command("ps", "-e")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	}

	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Subtract header line
	count := len(lines) - 1
	if count < 0 {
		count = 0
	}
	return count
}

// FormatPackageManagers formats package managers info
func (s *SoftwareInfo) FormatPackageManagers() string {
	if len(s.PackageManagers) == 0 {
		return "None detected"
	}

	var result []string
	for _, pm := range s.PackageManagers {
		result = append(result, fmt.Sprintf("%s (%d packages)", pm.Name, pm.Count))
	}
	return strings.Join(result, ", ")
}
