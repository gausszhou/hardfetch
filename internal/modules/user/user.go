package user

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

// UserInfo represents user information
type UserInfo struct {
	Username    string
	UserID      string
	GroupID     string
	Name        string
	HomeDir     string
	Shell       string
	Hostname    string
	Environment map[string]string
}

// GetUserInfo collects user information
func GetUserInfo() (*UserInfo, error) {
	info := &UserInfo{
		Environment: make(map[string]string),
	}

	// Get current user
	if currentUser, err := user.Current(); err == nil {
		info.Username = currentUser.Username
		info.UserID = currentUser.Uid
		info.GroupID = currentUser.Gid
		info.Name = currentUser.Name
		info.HomeDir = currentUser.HomeDir
	}

	// Get shell
	info.Shell = getShell()

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// Get important environment variables
	info.Environment = getEnvironment()

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

func getEnvironment() map[string]string {
	env := make(map[string]string)

	// Collect important environment variables
	importantVars := []string{
		"PATH", "HOME", "USER", "LOGNAME", "LANG", "LC_ALL", "LC_CTYPE",
		"TERM", "EDITOR", "VISUAL", "PAGER", "MANPAGER", "SHELL",
		"TZ", "TMPDIR", "TEMP", "TMP",
	}

	for _, key := range importantVars {
		if value := os.Getenv(key); value != "" {
			env[key] = value
		}
	}

	// Add platform-specific variables
	if runtime.GOOS == "windows" {
		if value := os.Getenv("USERPROFILE"); value != "" {
			env["USERPROFILE"] = value
		}
		if value := os.Getenv("APPDATA"); value != "" {
			env["APPDATA"] = value
		}
		if value := os.Getenv("LOCALAPPDATA"); value != "" {
			env["LOCALAPPDATA"] = value
		}
		if value := os.Getenv("ProgramFiles"); value != "" {
			env["ProgramFiles"] = value
		}
		if value := os.Getenv("ProgramFiles(x86)"); value != "" {
			env["ProgramFiles(x86)"] = value
		}
		if value := os.Getenv("SystemRoot"); value != "" {
			env["SystemRoot"] = value
		}
	}

	return env
}

// FormatEnvironment formats environment variables for display
func (u *UserInfo) FormatEnvironment() string {
	if len(u.Environment) == 0 {
		return "No environment variables found"
	}

	var builder strings.Builder
	for key, value := range u.Environment {
		// Truncate long values
		if len(value) > 50 {
			value = value[:47] + "..."
		}
		builder.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	return strings.TrimSpace(builder.String())
}

// FormatSummary formats a summary of user info
func (u *UserInfo) FormatSummary() string {
	var parts []string

	if u.Username != "" {
		parts = append(parts, fmt.Sprintf("User: %s", u.Username))
	}
	if u.Name != "" && u.Name != u.Username {
		parts = append(parts, fmt.Sprintf("Name: %s", u.Name))
	}
	if u.HomeDir != "" {
		parts = append(parts, fmt.Sprintf("Home: %s", u.HomeDir))
	}
	if u.Shell != "" {
		parts = append(parts, fmt.Sprintf("Shell: %s", u.Shell))
	}
	if u.Hostname != "" {
		parts = append(parts, fmt.Sprintf("Host: %s", u.Hostname))
	}

	return strings.Join(parts, " | ")
}
