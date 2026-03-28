package detect

import (
	"sync"

	"github.com/gausszhou/hardfetch/internal/detect/collector"
)

type Detector interface {
	Name() string
	Detect() (any, error)
}

type Result struct {
	System   *SystemInfo
	Hardware *HardwareInfo
	Network  *NetworkInfo
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
				if sys, ok := data.(*SystemInfo); ok {
					result.System = sys
				}
			case "hardware":
				if hw, ok := data.(*HardwareInfo); ok {
					result.Hardware = hw
				}
			case "network":
				if net, ok := data.(*NetworkInfo); ok {
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

func convertToSystemInfo(ci *collector.SystemInfoResult) *SystemInfo {
	if ci == nil {
		return &SystemInfo{}
	}
	return &SystemInfo{
		Hostname: ci.Hostname,
		OS:       ci.OSVersion,
		Host:     ci.Model,
		Kernel:   ci.Kernel,
		Shell:    ci.Shell,
		WM:       ci.WM,
		WMTheme:  ci.WMTheme,
		Theme:    ci.Theme,
		Font:     ci.Font,
		Cursor:   ci.Cursor,
		Terminal: ci.Terminal,
		Locale:   ci.Locale,
	}
}

func convertToHardwareInfo(hw *collector.HardwareResult, bat *collector.BatteryResult) *HardwareInfo {
	if hw == nil {
		return &HardwareInfo{}
	}
	info := &HardwareInfo{
		Memory: &MemoryInfo{
			Total: hw.Memory,
			Used:  hw.MemoryUsed,
		},
		Swap:  &SwapInfo{},
		Disks: make([]*DiskInfo, 0),
		GPUs:  make([]*GPUInfo, 0),
	}

	if hw.SwapTotal > 0 {
		info.Swap.Total = hw.SwapTotal
		info.Swap.Used = hw.SwapUsed
	}

	for _, d := range hw.Disks {
		info.Disks = append(info.Disks, &DiskInfo{
			Drive:      d.Drive,
			Total:      d.Total,
			Used:       d.Used,
			Free:       d.Free,
			FileSystem: d.FileSystem,
		})
	}

	for _, g := range hw.GPUs {
		info.GPUs = append(info.GPUs, &GPUInfo{
			Name:          g.Name,
			Vendor:        g.Vendor,
			VRAM:          g.VRAM,
			DriverVersion: g.DriverVersion,
		})
	}

	if bat != nil {
		info.Battery = &BatteryInfo{
			Percentage: bat.Percentage,
			Status:     bat.Status,
		}
	}

	return info
}

func convertToNetworkInfo(ni *collector.NetworkResult) *NetworkInfo {
	if ni == nil {
		return &NetworkInfo{}
	}
	info := &NetworkInfo{
		LocalIP: ni.LocalIP,
	}
	for _, iface := range ni.Interfaces {
		info.Interfaces = append(info.Interfaces, NetworkInterface{
			Name:      iface.Name,
			IPAddress: iface.IP,
		})
	}
	return info
}
