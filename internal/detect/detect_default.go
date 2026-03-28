//go:build !windows

package detect

import (
	"github.com/gausszhou/hardfetch/internal/modules/battery"
	"github.com/gausszhou/hardfetch/internal/modules/cpuinfo"
	"github.com/gausszhou/hardfetch/internal/modules/gpuinfo"
	"github.com/gausszhou/hardfetch/internal/modules/memory"
	"github.com/gausszhou/hardfetch/internal/modules/network"
)

func detectSystem() (any, error) {
	return &SystemInfo{}, nil
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
		Disks: []*DiskInfo{},
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
			Vendor:        g.Vendor,
			VRAM:          g.VRAM,
			DriverVersion: g.DriverVersion,
		})
	}
	return result
}
