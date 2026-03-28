package detect

import (
	"context"
	"sync"

	"github.com/gausszhou/hardfetch/internal/logger"
	"github.com/gausszhou/hardfetch/internal/modules/battery"
	"github.com/gausszhou/hardfetch/internal/modules/cpuinfo"
	"github.com/gausszhou/hardfetch/internal/modules/disk"
	"github.com/gausszhou/hardfetch/internal/modules/gpuinfo"
	"github.com/gausszhou/hardfetch/internal/modules/memory"
	"github.com/gausszhou/hardfetch/internal/modules/network"
	"github.com/gausszhou/hardfetch/internal/modules/sys"
)

func detectSystem() (any, error) {
	t := logger.StartTimer("system:detail")
	defer t.Stop()
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
	var mu sync.Mutex
	var wg sync.WaitGroup
	ctx := context.Background()

	cpu := &cpuinfo.Info{}
	mem := &memory.Info{}
	swap := &memory.SwapInfo{}
	gpus := []*gpuinfo.Info{}
	bat := &battery.Info{}
	disks := []*disk.Info{}

	detectFn := []struct {
		name string
		fn   func()
	}{
		{
			name: "cpu",
			fn: func() {
				defer wg.Done()
				t := logger.StartTimer("hardware:cpu")
				defer t.Stop()
				c, err := cpuinfo.Get()
				mu.Lock()
				if err == nil {
					*cpu = *c
				} else {
					logger.Debug(ctx, "cpu detect error", "error", err)
				}
				mu.Unlock()
			},
		},
		{
			name: "memory",
			fn: func() {
				defer wg.Done()
				t := logger.StartTimer("hardware:memory")
				defer t.Stop()
				m, s, err := memory.Get()
				mu.Lock()
				if err == nil {
					*mem = *m
					*swap = *s
				} else {
					logger.Debug(ctx, "memory detect error", "error", err)
				}
				mu.Unlock()
			},
		},
		{
			name: "gpu",
			fn: func() {
				defer wg.Done()
				t := logger.StartTimer("hardware:gpu")
				defer t.Stop()
				g, err := gpuinfo.Get()
				mu.Lock()
				if err == nil {
					gpus = g
				} else {
					logger.Debug(ctx, "gpu detect error", "error", err)
				}
				mu.Unlock()
			},
		},
		{
			name: "battery",
			fn: func() {
				defer wg.Done()
				t := logger.StartTimer("hardware:battery")
				defer t.Stop()
				b, err := battery.Get()
				mu.Lock()
				if err == nil {
					*bat = *b
				} else {
					logger.Debug(ctx, "battery detect error", "error", err)
				}
				mu.Unlock()
			},
		},
		{
			name: "disk",
			fn: func() {
				defer wg.Done()
				t := logger.StartTimer("hardware:disk")
				defer t.Stop()
				d, err := disk.Get()
				mu.Lock()
				if err == nil {
					disks = d
				} else {
					logger.Debug(ctx, "disk detect error", "error", err)
				}
				mu.Unlock()
			},
		},
	}

	wg.Add(len(detectFn))
	for _, df := range detectFn {
		go df.fn()
	}
	wg.Wait()

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
	t := logger.StartTimer("network:detail")
	defer t.Stop()
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
			VRAMString:    g.VRAMString,
			Frequency:     g.Frequency,
			Type:          g.Type,
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
