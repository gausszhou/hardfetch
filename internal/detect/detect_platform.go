package detect

import (
	"github.com/gausszhou/hardfetch/internal/modules/battery"
	"github.com/gausszhou/hardfetch/internal/modules/cpuinfo"
	"github.com/gausszhou/hardfetch/internal/modules/disk"
	"github.com/gausszhou/hardfetch/internal/modules/gpuinfo"
	"github.com/gausszhou/hardfetch/internal/modules/memory"
	"github.com/gausszhou/hardfetch/internal/modules/network"
	"github.com/gausszhou/hardfetch/internal/modules/sys"
)

func detectSystem() (any, error) {
	sysInfo, err := sys.Get()
	if err != nil {
		sysInfo = &sys.Info{}
	}

	return &SystemInfo{
		OS:       sysInfo.OS,
		Arch:     sysInfo.Arch,
		Kernel:   sysInfo.Kernel,
		Hostname: sysInfo.Hostname,
		Host:     sysInfo.Host,
		Uptime:   sysInfo.Uptime,
	}, nil
}

func detectHardware() (any, error) {
	cpu, err := cpuinfo.Get()
	if err != nil {
		cpu = &cpuinfo.Info{}
	}

	mem, swap, err := memory.Get()
	if err != nil {
		mem = &memory.Info{}
		swap = &memory.SwapInfo{}
	}

	gpus, err := gpuinfo.Get()
	if err != nil {
		gpus = []*gpuinfo.Info{}
	}

	bat, err := battery.Get()
	if err != nil {
		bat = &battery.Info{}
	}

	disks, err := disk.Get()
	if err != nil {
		disks = []*disk.Info{}
	}

	return &HardwareInfo{
		CPU: &CPUInfo{
			Model:        cpu.Model,
			Cores:        cpu.Cores,
			Threads:      cpu.Threads,
			Frequency:    cpu.Frequency,
			Architecture: cpu.Architecture,
		},
		Memory: &MemoryInfo{
			Total:     mem.Total,
			Used:      mem.Used,
			Available: mem.Available,
			Free:      mem.Free,
		},
		Swap: &SwapInfo{
			Total:     swap.Total,
			Used:      swap.Used,
			Free:      swap.Free,
			Available: swap.Available,
		},
		GPUs: convertGPUs(gpus),
		Battery: &BatteryInfo{
			Percentage:    bat.Percentage,
			Status:        bat.Status,
			TimeRemaining: bat.TimeRemaining,
		},
		Disks: convertDisks(disks),
	}, nil
}

func detectNetwork() (any, error) {
	net, err := network.Get()
	if err != nil {
		net = &network.Info{}
	}

	interfaces := make([]NetworkInterface, 0, len(net.Interfaces))
	for _, iface := range net.Interfaces {
		interfaces = append(interfaces, NetworkInterface{
			Name:       iface.Name,
			MACAddress: iface.MACAddress,
			IPAddress:  iface.IPAddress,
		})
	}

	return &NetworkInfo{
		Hostname:   net.Hostname,
		LocalIP:    net.LocalIP,
		PublicIP:   net.PublicIP,
		Interfaces: interfaces,
	}, nil
}

func convertGPUs(gpus []*gpuinfo.Info) []*GPUInfo {
	result := make([]*GPUInfo, 0, len(gpus))
	for _, g := range gpus {
		result = append(result, &GPUInfo{
			Name:          g.Name,
			VRAM:          g.VRAM,
			DriverVersion: g.DriverVersion,
		})
	}
	return result
}

func convertDisks(disks []*disk.Info) []*DiskInfo {
	result := make([]*DiskInfo, 0, len(disks))
	for _, d := range disks {
		result = append(result, &DiskInfo{
			Drive:      d.Drive,
			Total:      d.Total,
			Used:       d.Used,
			Free:       d.Free,
			FileSystem: d.FileSystem,
		})
	}
	return result
}
