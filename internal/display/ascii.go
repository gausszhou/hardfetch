package display

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

// FormatInfoWithColor formats with color codes
func FormatInfoWithColor(label, value string, colorCode string) string {
	const reset = "\033[0m"
	return colorCode + label + reset + ": " + value
}
