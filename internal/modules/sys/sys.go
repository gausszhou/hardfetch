package sys

import (
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v4/host"
)

type Info struct {
	OS       string
	Arch     string
	Kernel   string
	Hostname string
	Host     string
}

func Get() (*Info, error) {
	info := &Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: getHostname(),
	}

	platform, family, version, err := host.PlatformInformation()
	if err == nil {
		info.Host = platform
		if family != "" && family != platform {
			info.Host = family + " " + platform
		}
		if version != "" {
			info.Kernel = version
		}
	}

	if info.OS == "windows" {
		info.OS = "Windows"
	} else if info.OS == "darwin" {
		info.OS = "macOS"
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
