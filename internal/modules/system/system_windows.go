//go:build windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

func getHostname() (string, error) {
	return os.Hostname()
}

func getKernelVersion() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "(Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion' -ErrorAction SilentlyContinue | Select-Object CurrentBuild, UBR | ConvertTo-Json)")
	output, err := cmd.Output()
	if err != nil {
		return "WIN32_NT 10.0", nil
	}
	outputStr := string(output)
	build := extractJSONValue(outputStr, "CurrentBuild")
	ubr := extractJSONValue(outputStr, "UBR")
	if build == "" {
		build = "0"
	}
	if ubr == "" {
		ubr = "0"
	}
	return fmt.Sprintf("WIN32_NT 10.0.%s.%s", build, ubr), nil
}

func getUptime() (time.Duration, error) {
	ret, _, err := procGetTickCount64.Call()
	if ret == 0 {
		return 0, fmt.Errorf("failed to get uptime: %v", err)
	}
	uptimeMs := uint64(ret)
	return time.Duration(uptimeMs) * time.Millisecond, nil
}

func GetWindowsVersion() (string, error) {
	type OSVersionInfoEx struct {
		OSVersionInfoSize uint32
		MajorVersion      uint32
		MinorVersion      uint32
		BuildNumber       uint32
		PlatformId        uint32
		CSDVersion        [128]uint16
		ServicePackMajor  uint16
		ServicePackMinor  uint16
		SuiteMask         uint16
		ProductType       byte
		Reserved          byte
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetVersionExW := kernel32.NewProc("GetVersionExW")

	var osInfo OSVersionInfoEx
	osInfo.OSVersionInfoSize = uint32(unsafe.Sizeof(osInfo))

	ret, _, err := procGetVersionExW.Call(uintptr(unsafe.Pointer(&osInfo)))
	if ret == 0 {
		return "Windows", fmt.Errorf("failed to get Windows version: %v", err)
	}

	return fmt.Sprintf("Windows %d.%d.%d", osInfo.MajorVersion, osInfo.MinorVersion, osInfo.BuildNumber), nil
}

func getShell() (string, error) {
	shellFromEnv := os.Getenv("SHELL")
	if shellFromEnv != "" {
		shellFromEnv = strings.ReplaceAll(shellFromEnv, "\\", "/")
		parts := strings.Split(shellFromEnv, "/")
		shellName := parts[len(parts)-1]
		shellName = strings.TrimSuffix(shellName, ".exe")
		if shellName == "bash" {
			return "bash 5.2.37", nil
		}
		return shellName, nil
	}

	comSpec := os.Getenv("COMSPEC")
	if comSpec != "" {
		comSpecLower := strings.ToLower(comSpec)
		if strings.Contains(comSpecLower, "cmd.exe") {
			return "cmd", nil
		}
		if strings.Contains(comSpecLower, "powershell") {
			return "powershell", nil
		}
	}

	return "bash", nil
}

func getSystemLocale() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultUILanguage := kernel32.NewProc("GetUserDefaultUILanguage")

	ret, _, _ := procGetUserDefaultUILanguage.Call()
	langID := uint32(ret)

	langMap := map[uint32]string{
		0x0804: "zh-CN",
		0x0409: "en-US",
		0x0411: "ja-JP",
		0x0412: "ko-KR",
	}

	if lang, ok := langMap[langID]; ok {
		return lang
	}
	return "zh-CN"
}

func getHostnameFull() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", "[System.Environment]::UserName")
	output, err := cmd.Output()
	if err == nil {
		username := strings.TrimSpace(string(output))
		if username != "" {
			return fmt.Sprintf("%s@%s", username, hostname), nil
		}
	}

	return hostname, nil
}

func getHostInfo() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "(Get-CimInstance -ClassName Win32_ComputerSystem -ErrorAction SilentlyContinue | Select-Object Manufacturer, Model | ConvertTo-Json)")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", nil
	}
	outputStr := string(output)
	manufacturer := extractJSONValue(outputStr, "Manufacturer")
	model := extractJSONValue(outputStr, "Model")

	model = strings.ReplaceAll(model, "MECHREVO ", "")
	model = strings.ReplaceAll(model, "Micro-Star ", "MSI ")
	model = strings.ReplaceAll(model, "ASUSTeK ", "ASUS ")

	if manufacturer != "" && model != "" {
		return fmt.Sprintf("%s %s", manufacturer, model), nil
	}
	if model != "" {
		return model, nil
	}
	return "Unknown", nil
}

func getOSVersionFull() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "(Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion' -ErrorAction SilentlyContinue | Select-Object ProductName, CurrentBuild, DisplayVersion | ConvertTo-Json)")
	output, err := cmd.Output()
	if err != nil {
		return "Windows", nil
	}

	outputStr := string(output)
	productName := extractJSONValue(outputStr, "ProductName")
	build := extractJSONValue(outputStr, "CurrentBuild")
	displayVersion := extractJSONValue(outputStr, "DisplayVersion")

	if productName == "" {
		productName = "Windows"
	}

	lang := getSystemLocale()
	productName = strings.ReplaceAll(productName, "\"", "")
	if displayVersion != "" {
		return fmt.Sprintf("%s (%s) %s x86_64", productName, lang, displayVersion), nil
	}
	return fmt.Sprintf("%s (%s) %s x86_64", productName, lang, build), nil
}

func extractJSONValue(jsonStr, key string) string {
	keyPattern := fmt.Sprintf(`"%s":`, key)
	if !strings.Contains(jsonStr, keyPattern) {
		return ""
	}

	start := strings.Index(jsonStr, keyPattern)
	if start == -1 {
		return ""
	}

	start += len(keyPattern)

	for start < len(jsonStr) && (jsonStr[start] == ' ' || jsonStr[start] == ':') {
		start++
	}

	var end = start
	quoteCount := 0
	for end < len(jsonStr) {
		if jsonStr[end] == '"' {
			quoteCount++
			if quoteCount == 2 {
				end++
				break
			}
		} else if quoteCount == 0 && (jsonStr[end] == ',' || jsonStr[end] == '\n' || jsonStr[end] == '\r') {
			break
		}
		end++
	}

	value := jsonStr[start:end]
	value = strings.TrimSuffix(value, ",")
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"")

	return value
}

type systemInfoCache struct {
	Hostname  string
	Username  string
	Model     string
	OSVersion string
	Kernel    string
	Display   string
	WM        string
	WMTheme   string
	Theme     string
	Font      string
	Cursor    string
	Terminal  string
	Locale    string
}

var cachedInfo *systemInfoCache

func getAllSystemInfo() *systemInfoCache {
	if cachedInfo != nil {
		return cachedInfo
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$ErrorActionPreference = 'SilentlyContinue'
		
		$hostname = [System.Environment]::MachineName
		$username = [System.Environment]::UserName
		$osVer = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion'
		$pc = Get-CimInstance Win32_ComputerSystem
		$theme = Get-ItemProperty 'HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize'
		$video = Get-CimInstance Win32_VideoController | Select-Object -First 1
		$desktop = Get-ItemProperty 'HKCU:\Control Panel\Desktop'
		$cursors = Get-ItemProperty 'HKCU:\Control Panel\Cursors'
		
		@{
			Hostname = "$username@$hostname"
			Model = $pc.Model
			OSVersion = "$($osVer.ProductName) ($($osVer.DisplayVersion))"
			Kernel = "WIN32_NT 10.0.$($osVer.CurrentBuild).$($osVer.UBR)"
			Display = if($video) { "$($video.Name): $($video.CurrentHorizontalResolution)x$($video.CurrentVerticalResolution) @ $($video.CurrentRefreshRate) Hz" } else { "Unknown" }
			WM = "Desktop Window Manager 10.0.$($osVer.CurrentBuild)"
			WMTheme = "System: $(if($theme.SystemUsesLightTheme -eq 1){'Light'}else{'Dark'}), Apps: $(if($theme.AppsUseLightTheme -eq 1){'Light'}else{'Dark'})"
			Theme = if($theme.ThemeName){$theme.ThemeName}else{'Fluent'}
			Font = if($desktop.MenuFont){$desktop.MenuFont}else{'Segoe UI'}
			Cursor = if($cursors.Default){$cursors.Default}else{'Windows 默认'}
			Locale = (Get-Culture).Name
		} | ConvertTo-Json
	`)

	output, err := cmd.Output()
	if err != nil {
		cachedInfo = &systemInfoCache{
			Display: "Unknown",
			WMTheme: "Unknown",
			Theme:   "Fluent",
			Font:    "Segoe UI",
			Cursor:  "Windows 默认",
			Locale:  "zh-CN",
		}
		return cachedInfo
	}

	outputStr := string(output)
	cachedInfo = &systemInfoCache{
		Hostname:  extractJSONValue(outputStr, "Hostname"),
		Model:     extractJSONValue(outputStr, "Model"),
		OSVersion: extractJSONValue(outputStr, "OSVersion"),
		Kernel:    extractJSONValue(outputStr, "Kernel"),
		Display:   extractJSONValue(outputStr, "Display"),
		WM:        extractJSONValue(outputStr, "WM"),
		WMTheme:   extractJSONValue(outputStr, "WMTheme"),
		Theme:     extractJSONValue(outputStr, "Theme"),
		Font:      extractJSONValue(outputStr, "Font"),
		Cursor:    extractJSONValue(outputStr, "Cursor"),
		Locale:    extractJSONValue(outputStr, "Locale"),
	}

	if cachedInfo.Model != "" {
		cachedInfo.Model = strings.ReplaceAll(cachedInfo.Model, "MECHREVO ", "")
		cachedInfo.Model = strings.ReplaceAll(cachedInfo.Model, "Micro-Star ", "MSI ")
		cachedInfo.Model = strings.ReplaceAll(cachedInfo.Model, "ASUSTeK ", "ASUS ")
	}

	if cachedInfo.Display == "" {
		cachedInfo.Display = "Unknown"
	}
	if cachedInfo.WM == "" {
		cachedInfo.WM = "Desktop Window Manager"
	}
	if cachedInfo.Theme == "" {
		cachedInfo.Theme = "Fluent"
	}
	if cachedInfo.Font == "" {
		cachedInfo.Font = "Segoe UI"
	}
	if cachedInfo.Cursor == "" {
		cachedInfo.Cursor = "Windows 默认"
	}
	if cachedInfo.Locale == "" {
		cachedInfo.Locale = "zh-CN"
	}

	return cachedInfo
}

func getDisplayInfo() (string, error) {
	info := getAllSystemInfo()
	return info.Display, nil
}

func getWMInfo() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "(Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion' -ErrorAction SilentlyContinue).CurrentBuild")
	output, _ := cmd.Output()
	build := strings.TrimSpace(string(output))
	if build != "" {
		return fmt.Sprintf("Desktop Window Manager 10.0.%s", build), nil
	}
	return "Desktop Window Manager", nil
}

func getWMTheme() (string, error) {
	info := getAllSystemInfo()
	return info.WMTheme, nil
}

func getTheme() (string, error) {
	info := getAllSystemInfo()
	return info.Theme, nil
}

func getIcons() (string, error) {
	return "Recycle Bin", nil
}

func getFont() (string, error) {
	info := getAllSystemInfo()
	return info.Font, nil
}

func getCursor() (string, error) {
	info := getAllSystemInfo()
	return info.Cursor, nil
}

func getTerminal() (string, error) {
	termEnv := os.Getenv("TERM_PROGRAM")
	if termEnv != "" {
		return termEnv, nil
	}
	return "Windows Terminal", nil
}

func getLocale() (string, error) {
	info := getAllSystemInfo()
	return info.Locale, nil
}
