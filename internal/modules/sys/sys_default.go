//go:build !windows

package sys

import (
	"os"
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
	info := &Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: getHostname(),
	}

	platform, family, version, _ := host.PlatformInformation()
	if platform != "" {
		info.Host = platform
		if family != "" && family != platform {
			info.Host = family + " " + platform
		}
		if version != "" && runtime.GOOS != "darwin" {
			info.Kernel = version
		}
	}

	switch runtime.GOOS {
	case "darwin":
		info.Host = getDarwinHost()
		info.Kernel = getDarwinKernel()
		info.OS = "macOS"
	case "linux":
		info.Kernel = getLinuxKernel()
		info.OS = "Linux"
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

func getLinuxKernel() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
