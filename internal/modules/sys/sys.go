package sys

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v4/host"
	"golang.org/x/sys/windows"
)

type Info struct {
	OS       string
	Arch     string
	Kernel   string
	Hostname string
	Host     string
	Uptime   time.Duration
	Shell    string
}

func Get() (*Info, error) {
	info := &Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: getHostname(),
	}

	switch runtime.GOOS {
	case "windows":
		info.OS = getWindowsOS()
		info.Host = getWindowsHost()
		info.Kernel = getWindowsKernel()
		info.Shell = getWindowsShell()
	case "darwin":
		info.Host = getDarwinHost()
		info.Kernel = getDarwinKernel()
		info.Shell = getDarwinShell()
	case "linux":
		info.Host = getLinuxHost()
		info.Kernel = getLinuxKernel()
		info.Shell = getLinuxShell()
	}

	uptime, err := host.BootTime()
	if err == nil {
		info.Uptime = time.Duration(time.Now().Unix()-int64(uptime)) * time.Second
	}

	return info, nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}
	return hostname
}

func getWindowsOS() string {
	var hKey windows.Handle
	keyPath, _ := windows.UTF16PtrFromString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, keyPath, 0, windows.KEY_READ|windows.KEY_WOW64_64KEY, &hKey); err != nil {
		return "Windows"
	}
	defer windows.RegCloseKey(hKey)

	productName := queryRegString(hKey, "ProductName")
	displayVersion := queryRegString(hKey, "DisplayVersion")
	build := queryRegString(hKey, "CurrentBuild")

	osStr := "Windows"
	if productName != "" {
		osStr = productName
	}
	if displayVersion != "" {
		osStr += " " + displayVersion
	}
	if build != "" {
		osStr += " (" + build + ")"
	}
	osStr += " " + runtime.GOARCH

	return osStr
}

func getWindowsHost() string {
	var hKey windows.Handle
	keyPath, _ := windows.UTF16PtrFromString(`SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName`)
	if err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, keyPath, 0, windows.KEY_READ|windows.KEY_WOW64_64KEY, &hKey); err != nil {
		return ""
	}
	defer windows.RegCloseKey(hKey)

	return queryRegString(hKey, "ComputerName")
}

func getWindowsKernel() string {
	var hKey windows.Handle
	keyPath, _ := windows.UTF16PtrFromString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, keyPath, 0, windows.KEY_READ|windows.KEY_WOW64_64KEY, &hKey); err != nil {
		return ""
	}
	defer windows.RegCloseKey(hKey)

	build := queryRegString(hKey, "CurrentBuild")
	ubr := queryRegString(hKey, "UBR")

	kernel := "WIN32_NT 10.0"
	if build != "" {
		kernel += "." + build
	}
	if ubr != "" {
		kernel += "." + ubr
	}

	return kernel
}

func getWindowsShell() string {
	comspec := os.Getenv("COMSPEC")
	if comspec != "" {
		parts := strings.Split(comspec, "\\")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	shell := os.Getenv("SHELL")
	if shell != "" {
		parts := strings.Split(shell, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return "cmd"
}

func queryRegString(hKey windows.Handle, name string) string {
	namePtr, _ := windows.UTF16PtrFromString(name)
	var buf [64]byte
	var bufType uint32
	size := uint32(len(buf))
	err := windows.RegQueryValueEx(hKey, namePtr, nil, &bufType, &buf[0], &size)
	if err != nil {
		return ""
	}
	if bufType == windows.REG_DWORD {
		val := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
		return fmt.Sprintf("%d", val)
	}
	if size > 0 && size/2 < uint32(len(buf)) {
		return windows.UTF16ToString((*[32]uint16)(unsafe.Pointer(&buf[0]))[:size/2])
	}
	return ""
}

func getDarwinHost() string {
	platform, family, _, _ := host.PlatformInformation()
	if platform != "" {
		if family != "" && family != platform {
			return family + " " + platform
		}
		return platform
	}
	return "macOS"
}

func getDarwinKernel() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getDarwinShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		parts := strings.Split(shell, "/")
		if len(parts) > 0 {
			shell = parts[len(parts)-1]
		}
	}
	return shell
}

func getLinuxHost() string {
	platform, family, version, _ := host.PlatformInformation()
	if platform != "" {
		if family != "" && family != platform {
			return family + " " + platform
		}
		if version != "" {
			return platform + " " + version
		}
		return platform
	}
	return "Linux"
}

func getLinuxKernel() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getLinuxShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		parts := strings.Split(shell, "/")
		if len(parts) > 0 {
			shell = parts[len(parts)-1]
		}
	}
	return shell
}
