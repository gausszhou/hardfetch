package display

import (
	"fmt"
	"runtime"
	"strings"
)

// Logo represents an ASCII logo
type Logo struct {
	Lines []string
	Color string
}

// GetLogo returns the appropriate ASCII logo for the current OS
func GetLogo() *Logo {
	switch runtime.GOOS {
	case "windows":
		return getWindowsLogo()
	case "linux":
		return getLinuxLogo()
	case "darwin":
		return getMacOSLogo()
	default:
		return getGenericLogo()
	}
}

func getWindowsLogo() *Logo {
	return &Logo{
		Lines: []string{
			"",
			"    .-^-.",
			"   /     \\",
			"  |       |",
			"  |       |",
			"   \\     /",
			"    '-.-'",
			"      |",
			"    .-'-.",
			"   /     \\",
			"  |       |",
			"  |       |",
			"   \\     /",
			"    '-.-'",
			"",
		},
		Color: "blue",
	}
}

func getLinuxLogo() *Logo {
	return &Logo{
		Lines: []string{
			"",
			"       __",
			"      /  \\",
			"     /    \\",
			"    /  /\\  \\",
			"   /  /  \\  \\",
			"  /  /    \\  \\",
			" /  /      \\  \\",
			"/__/        \\__\\",
			"",
		},
		Color: "green",
	}
}

func getMacOSLogo() *Logo {
	return &Logo{
		Lines: []string{
			"",
			"      .:'",
			"    __ :'__",
			" .'`__`-'__``.",
			":__________.-'",
			":_________:",
			" :________`-;",
			"  `.__.-.__.'",
			"",
		},
		Color: "cyan",
	}
}

func getGenericLogo() *Logo {
	return &Logo{
		Lines: []string{
			"",
			"  ██████╗ ██╗   ██╗███████╗",
			" ██╔════╝ ██║   ██║██╔════╝",
			" ██║  ███╗██║   ██║█████╗  ",
			" ██║   ██║██║   ██║██╔══╝  ",
			" ╚██████╔╝╚██████╔╝███████╗",
			"  ╚═════╝  ╚═════╝ ╚══════╝",
			"",
		},
		Color: "yellow",
	}
}

// PrintLogo prints the ASCII logo
func PrintLogo(logo *Logo) {
	for _, line := range logo.Lines {
		fmt.Println(line)
	}
}

// FormatInfo formats system information for display
func FormatInfo(label, value string, color string) string {
	// Apply color if supported
	// For now, just return formatted string
	return fmt.Sprintf("%-15s: %s", label, value)
}

// FormatInfoWithColor formats with color codes
func FormatInfoWithColor(label, value string, colorCode string) string {
	const reset = "\033[0m"
	return fmt.Sprintf("%s%-15s%s: %s", colorCode, label, reset, value)
}

// GetColorCode returns ANSI color code
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

// FormatTable formats information in a table
func FormatTable(data map[string]string, maxLabelWidth int) string {
	var builder strings.Builder

	for label, value := range data {
		padding := maxLabelWidth - len(label)
		if padding < 0 {
			padding = 0
		}
		builder.WriteString(label)
		builder.WriteString(strings.Repeat(" ", padding))
		builder.WriteString(": ")
		builder.WriteString(value)
		builder.WriteString("\n")
	}

	return builder.String()
}
