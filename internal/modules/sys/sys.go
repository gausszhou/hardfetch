package sys

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

type Info struct {
	OS       string
	Arch     string
	Kernel   string
	Hostname string
	Host     string
	Uptime   time.Duration
}

func Get() (*Info, error) {
	hostInfo, err := host.Info()
	if err != nil {
		hostInfo = &host.InfoStat{}
	}

	info := &Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hostInfo.Hostname,
	}

	switch runtime.GOOS {
	case "windows":
		info.OS = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
		info.Host = getWindowsHost()
		info.Kernel = hostInfo.KernelVersion
	case "darwin":
		info.Host = getDarwinHost()
		info.Kernel = getDarwinKernel()
	case "linux":
		info.Host = getLinuxHost()
		info.Kernel = getLinuxKernel()
	}

	uptime, err := host.BootTime()
	if err == nil {
		info.Uptime = time.Duration(time.Now().Unix()-int64(uptime)) * time.Second
	}

	return info, nil
}

func getWindowsHost() string {
	script := `(Get-CimInstance Win32_ComputerSystem).Model`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
