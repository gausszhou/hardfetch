package cpuinfo

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v4/cpu"
)

type Info struct {
	Model        string
	Cores        int
	Threads      int
	Frequency    string
	Architecture string
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

	return info, nil
}
