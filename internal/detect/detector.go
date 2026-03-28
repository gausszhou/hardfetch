package detect

import (
	"sync"

	"github.com/gausszhou/hardfetch/internal/modules/collector"
	"github.com/gausszhou/hardfetch/internal/modules/hardware"
	"github.com/gausszhou/hardfetch/internal/modules/network"
	"github.com/gausszhou/hardfetch/internal/modules/system"
)

type Detector interface {
	Name() string
	Detect() (any, error)
}

type Result struct {
	System   *system.SystemInfo
	Hardware *hardware.HardwareInfo
	Network  *network.NetworkInfo
}

type coreDetector struct {
	name string
	fn   func() (any, error)
}

func (d *coreDetector) Name() string {
	return d.name
}

func (d *coreDetector) Detect() (any, error) {
	return d.fn()
}

var (
	result     *Result
	resultOnce sync.Once
)

func GetCoreDetectors() []Detector {
	return []Detector{
		&coreDetector{name: "system", fn: detectSystem},
		&coreDetector{name: "hardware", fn: detectHardware},
		&coreDetector{name: "network", fn: detectNetwork},
	}
}

func Detect(detectors ...Detector) *Result {
	resultOnce.Do(func() {
		result = &Result{}
		collectAll(detectors)
	})
	return result
}

func collectAll(detectors []Detector) {
	var wg sync.WaitGroup
	wg.Add(len(detectors))

	for _, d := range detectors {
		go func(detector Detector) {
			defer wg.Done()
			data, err := detector.Detect()
			if err != nil {
				return
			}
			switch detector.Name() {
			case "system":
				if sys, ok := data.(*system.SystemInfo); ok {
					result.System = sys
				}
			case "hardware":
				if hw, ok := data.(*hardware.HardwareInfo); ok {
					result.Hardware = hw
				}
			case "network":
				if net, ok := data.(*network.NetworkInfo); ok {
					result.Network = net
				}
			}
		}(d)
	}

	wg.Wait()
}

func detectSystem() (any, error) {
	c := collector.GetCollector()
	return convertToSystemInfo(c.SystemInfo), nil
}

func detectHardware() (any, error) {
	c := collector.GetCollector()
	return convertToHardwareInfo(c.Hardware, c.Battery), nil
}

func detectNetwork() (any, error) {
	c := collector.GetCollector()
	return convertToNetworkInfo(c.Network), nil
}

func convertToSystemInfo(ci *collector.SystemInfoResult) *system.SystemInfo {
	if ci == nil {
		return &system.SystemInfo{}
	}
	return &system.SystemInfo{
		Hostname: ci.Hostname,
		OS:       ci.OSVersion,
		Host:     ci.Model,
		Kernel:   ci.Kernel,
		Shell:    ci.Shell,
		Display:  ci.Display,
		WM:       ci.WM,
		WMTheme:  ci.WMTheme,
		Theme:    ci.Theme,
		Font:     ci.Font,
		Cursor:   ci.Cursor,
		Terminal: ci.Terminal,
		Locale:   ci.Locale,
	}
}

func convertToHardwareInfo(hw *collector.HardwareResult, bat *collector.BatteryResult) *hardware.HardwareInfo {
	if hw == nil {
		return &hardware.HardwareInfo{}
	}
	info := &hardware.HardwareInfo{
		Memory: &hardware.MemoryInfo{
			Total: hw.Memory,
			Used:  hw.MemoryUsed,
		},
		Swap:  &hardware.SwapInfo{},
		Disks: make([]*hardware.DiskInfo, 0),
		GPUs:  make([]*hardware.GPUInfo, 0),
	}

	if hw.SwapTotal > 0 {
		info.Swap.Total = hw.SwapTotal
		info.Swap.Used = hw.SwapUsed
	}

	for _, d := range hw.Disks {
		info.Disks = append(info.Disks, &hardware.DiskInfo{
			Drive:      d.Drive,
			Total:      d.Total,
			Used:       d.Used,
			Free:       d.Free,
			FileSystem: d.FileSystem,
		})
	}

	for _, g := range hw.GPUs {
		info.GPUs = append(info.GPUs, &hardware.GPUInfo{
			Name:          g.Name,
			Vendor:        g.Vendor,
			VRAM:          g.VRAM,
			DriverVersion: g.DriverVersion,
		})
	}

	if bat != nil {
		info.Battery = &hardware.BatteryInfo{
			Percentage: bat.Percentage,
			Status:     bat.Status,
		}
	}

	return info
}

func convertToNetworkInfo(ni *collector.NetworkResult) *network.NetworkInfo {
	if ni == nil {
		return &network.NetworkInfo{}
	}
	info := &network.NetworkInfo{
		LocalIP: ni.LocalIP,
	}
	for _, iface := range ni.Interfaces {
		info.Interfaces = append(info.Interfaces, network.NetworkInterface{
			Name:      iface.Name,
			IPAddress: iface.IP,
		})
	}
	return info
}
