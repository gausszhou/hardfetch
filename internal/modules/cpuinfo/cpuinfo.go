package cpuinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
)

type Info struct {
	Model        string
	Cores        int
	Threads      int
	Frequency    string
	Architecture string
	Uptime       time.Duration
}

func Get() (*Info, error) {
	info := &Info{
		Architecture: runtime.GOARCH,
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.Model = cpuInfo[0].ModelName
		info.Cores = int(cpuInfo[0].Cores)
		info.Threads = runtime.NumCPU()

		if cpuInfo[0].Mhz > 0 {
			info.Frequency = fmt.Sprintf("%.2f GHz", cpuInfo[0].Mhz/1000)
		}
	}

	uptime, err := host.BootTime()
	if err == nil {
		info.Uptime = time.Duration(time.Now().Unix()-int64(uptime)) * time.Second
	}

	return info, nil
}
