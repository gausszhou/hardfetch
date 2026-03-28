package sys

import (
	"os"
	"runtime"
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

	uptime, err := host.BootTime()
	if err == nil {
		info.Uptime = time.Duration(time.Now().Unix()-int64(uptime)) * time.Second
	}

	switch info.OS {
	case "windows":
		info.OS = "Windows"
	case "darwin":
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
