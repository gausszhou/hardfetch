//go:build windows

package detect

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"sync"

	"github.com/gausszhou/hardfetch/internal/logger"
)

type systemInfoResult struct {
	Hostname   string `json:"Hostname"`
	Model      string `json:"Model"`
	OSVersion  string `json:"OSVersion"`
	Kernel     string `json:"Kernel"`
	Display    string `json:"Display"`
	WM         string `json:"WM"`
	WMTheme    string `json:"WMTheme"`
	Theme      string `json:"Theme"`
	Font       string `json:"Font"`
	Cursor     string `json:"Cursor"`
	Locale     string `json:"Locale"`
	Shell      string `json:"Shell"`
	Terminal   string `json:"Terminal"`
	Uptime     uint64 `json:"Uptime"`
	Processors int    `json:"Processors"`
}

type gpuResult struct {
	Name          string `json:"Name"`
	Vendor        string `json:"Vendor"`
	VRAM          uint64 `json:"VRAM"`
	DriverVersion string `json:"DriverVersion"`
}

type hardwareResult struct {
	Memory      uint64       `json:"Memory"`
	MemoryUsed  uint64       `json:"MemoryUsed"`
	SwapTotal   uint64       `json:"SwapTotal"`
	SwapUsed    uint64       `json:"SwapUsed"`
	GPUs        []gpuResult  `json:"GPUs"`
	Disks       []diskResult `json:"Disks"`
	MonitorName string       `json:"MonitorName"`
}

type diskResult struct {
	Drive      string `json:"Drive"`
	Total      uint64 `json:"Total"`
	Used       uint64 `json:"Used"`
	Free       uint64 `json:"Free"`
	FileSystem string `json:"FileSystem"`
}

type batteryResult struct {
	Percentage int    `json:"Percentage"`
	Status     string `json:"Status"`
}

type networkResult struct {
	LocalIP    string            `json:"LocalIP"`
	Interfaces []interfaceResult `json:"Interfaces"`
}

type interfaceResult struct {
	Name string `json:"Name"`
	IP   string `json:"IP"`
}

var (
	collectorInstance *windowsCollector
	collectorOnce     sync.Once
)

type windowsCollector struct {
	mu         sync.RWMutex
	SystemInfo *systemInfoResult
	Hardware   *hardwareResult
	Battery    *batteryResult
	Network    *networkResult
}

func GetCollector() *windowsCollector {
	collectorOnce.Do(func() {
		collectorInstance = &windowsCollector{}
		collectorInstance.collectAll()
	})
	return collectorInstance
}

func (c *windowsCollector) collectAll() {
	var wg sync.WaitGroup

	wg.Add(4)
	go func() {
		defer wg.Done()
		t := logger.StartTimer("collectSystemInfo")
		logger.Debug(context.Background(), "executing powershell command", "command_type", "system_info")
		c.SystemInfo = c.collectSystemInfo()
		t.Stop()
	}()
	go func() {
		defer wg.Done()
		t := logger.StartTimer("collectHardware")
		logger.Debug(context.Background(), "executing powershell command", "command_type", "hardware")
		c.Hardware = c.collectHardware()
		t.Stop()
	}()
	go func() {
		defer wg.Done()
		t := logger.StartTimer("collectBattery")
		logger.Debug(context.Background(), "executing powershell command", "command_type", "battery")
		c.Battery = c.collectBattery()
		t.Stop()
	}()
	go func() {
		defer wg.Done()
		t := logger.StartTimer("collectNetwork")
		logger.Debug(context.Background(), "executing powershell command", "command_type", "network")
		c.Network = c.collectNetwork()
		t.Stop()
	}()

	wg.Wait()
}

func (c *windowsCollector) collectSystemInfo() *systemInfoResult {
	result := &systemInfoResult{}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$ErrorActionPreference = 'SilentlyContinue'
		
		$hostname = [System.Environment]::MachineName
		$username = [System.Environment]::UserName
		$osVer = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion'
		$pc = Get-CimInstance Win32_ComputerSystem
		$theme = Get-ItemProperty 'HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize'
		$video = Get-CimInstance Win32_VideoController | Select-Object -First 1
		$desktop = Get-ItemProperty 'HKCU:\Control Panel\Desktop'
		$cursors = Get-ItemProperty 'HKCU:\Control Panel\Cursors'
		
		$shell = if($env:SHELL) { $env:SHELL } elseif($env:COMSPEC) { $env:COMSPEC } else { 'cmd' }
		$shell = [System.IO.Path]::GetFileNameWithoutExtension($shell)
		
		$uptime = (Get-Date) - (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
		$uptimeSec = [int]$uptime.TotalSeconds
		
		$proc = Get-CimInstance Win32_Processor | Measure-Object -Property NumberOfLogicalProcessors -Sum
		$processors = [int]$proc.Sum
		
		$term = if($env:TERM_PROGRAM) { $env:TERM_PROGRAM } else { 'Windows Terminal' }
		
		@{
			Hostname = "$username@$hostname"
			Model = $pc.Model
			OSVersion = "$($osVer.ProductName) ($($osVer.DisplayVersion))"
			Kernel = "WIN32_NT 10.0.$($osVer.CurrentBuild).$($osVer.UBR)"
			Display = if($video) { "$($video.Name): $($video.CurrentHorizontalResolution)x$($video.CurrentVerticalResolution) @ $($video.CurrentRefreshRate) Hz" } else { "Unknown" }
			WM = "Desktop Window Manager 10.0.$($osVer.CurrentBuild)"
			WMTheme = "System: $(if($theme.SystemUsesLightTheme -eq 1){'Light'}else{'Dark'}), Apps: $(if($theme.AppsUseLightTheme -eq 1){'Light'}else{'Dark'})"
			Theme = if($theme.ThemeName){$theme.ThemeName}else{'Fluent'}
			Font = if($desktop.MenuFont){$desktop.MenuFont}else{'Segoe UI'}
			Cursor = if($cursors.Default){$cursors.Default}else{'Windows 默认'}
			Locale = (Get-Culture).Name
			Shell = $shell
			Terminal = $term
			Uptime = $uptimeSec
			Processors = $processors
		} | ConvertTo-Json -Compress
	`)

	output, err := cmd.Output()
	if err != nil {
		return result
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return result
	}

	json.Unmarshal([]byte(outputStr), result)

	result.Model = strings.ReplaceAll(result.Model, "MECHREVO ", "")
	result.Model = strings.ReplaceAll(result.Model, "Micro-Star ", "MSI ")
	result.Model = strings.ReplaceAll(result.Model, "ASUSTeK ", "ASUS ")

	if result.Shell != "" {
		result.Shell = strings.ReplaceAll(result.Shell, "C:\\Windows\\System32\\", "")
		result.Shell = strings.ReplaceAll(result.Shell, ".exe", "")
	}

	return result
}

func (c *windowsCollector) collectHardware() *hardwareResult {
	result := &hardwareResult{
		GPUs:  make([]gpuResult, 0),
		Disks: make([]diskResult, 0),
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$ErrorActionPreference = 'SilentlyContinue'
		
		$mem = Get-CimInstance Win32_OperatingSystem
		$memTotal = [int64]$mem.TotalVisibleMemorySize * 1024
		$memFree = [int64]$mem.FreePhysicalMemory * 1024
		$memUsed = $memTotal - $memFree
		
		$pf = Get-CimInstance Win32_PageFileUsage | Select-Object -First 1
		$swapTotal = [int64]0
		$swapUsed = [int64]0
		if ($pf -ne $null) {
			$swapTotal = [int64]$pf.AllocatedBaseSize * 1024 * 1024
			$swapUsed = [int64]$pf.CurrentUsage * 1024 * 1024
		}
		
		$gpus = Get-CimInstance Win32_VideoController | ForEach-Object {
			@{
				Name = $_.Name
				Vendor = $_.AdapterCompatibility
				VRAM = [int64]$_.AdapterRAM
				DriverVersion = $_.DriverVersion
			}
		}
		
		$disks = Get-WmiObject Win32_LogicalDisk | Where-Object {$_.DriveType -eq 3} | ForEach-Object {
			@{
				Drive = $_.DeviceID
				Total = [int64]$_.Size
				Free = [int64]$_.FreeSpace
				Used = [int64]($_.Size - $_.FreeSpace)
				FileSystem = $_.FileSystem
			}
		}
		
		$monitor = Get-CimInstance WmiMonitorID -Namespace root/wmi -ErrorAction SilentlyContinue | Select-Object -First 1
		$monitorName = ""
		if ($monitor -ne $null) {
			$monitorName = [System.Text.Encoding]::ASCII.GetString($monitor.UserFriendlyName)
		}
		
		@{
			Memory = $memTotal
			MemoryUsed = $memUsed
			SwapTotal = $swapTotal
			SwapUsed = $swapUsed
			GPUs = $gpus
			Disks = $disks
			MonitorName = $monitorName
		} | ConvertTo-Json -Compress -Depth 3
	`)

	output, err := cmd.Output()
	if err == nil {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" && outputStr != "{}" {
			var hwData struct {
				GPUs        interface{} `json:"GPUs"`
				Disks       interface{} `json:"Disks"`
				Memory      uint64      `json:"Memory"`
				MemoryUsed  uint64      `json:"MemoryUsed"`
				SwapTotal   uint64      `json:"SwapTotal"`
				SwapUsed    uint64      `json:"SwapUsed"`
				MonitorName string      `json:"MonitorName"`
			}
			if err := json.Unmarshal([]byte(outputStr), &hwData); err != nil {
			} else {
				result.Memory = hwData.Memory
				result.MemoryUsed = hwData.MemoryUsed
				result.SwapTotal = hwData.SwapTotal
				result.SwapUsed = hwData.SwapUsed
				result.MonitorName = hwData.MonitorName

				if gpus, ok := hwData.GPUs.([]interface{}); ok {
					for _, g := range gpus {
						if gm, ok := g.(map[string]interface{}); ok {
							gpu := gpuResult{}
							if v, ok := gm["Name"].(string); ok {
								gpu.Name = v
							}
							if v, ok := gm["Vendor"].(string); ok {
								gpu.Vendor = v
							}
							if v, ok := gm["VRAM"].(float64); ok {
								gpu.VRAM = uint64(v)
							}
							if v, ok := gm["DriverVersion"].(string); ok {
								gpu.DriverVersion = v
							}
							result.GPUs = append(result.GPUs, gpu)
						}
					}
				}

				if disks, ok := hwData.Disks.([]interface{}); ok {
					for _, d := range disks {
						if dm, ok := d.(map[string]interface{}); ok {
							disk := diskResult{}
							if v, ok := dm["Drive"].(string); ok {
								disk.Drive = v
							}
							if v, ok := dm["Total"].(float64); ok {
								disk.Total = uint64(v)
							}
							if v, ok := dm["Used"].(float64); ok {
								disk.Used = uint64(v)
							}
							if v, ok := dm["Free"].(float64); ok {
								disk.Free = uint64(v)
							}
							if v, ok := dm["FileSystem"].(string); ok {
								disk.FileSystem = v
							}
							result.Disks = append(result.Disks, disk)
						}
					}
				}
			}
		}
	}

	for i := range result.GPUs {
		if result.GPUs[i].Vendor == "" {
			result.GPUs[i].Vendor = detectVendor(result.GPUs[i].Name)
		}
	}

	return result
}

func (c *windowsCollector) collectBattery() *batteryResult {
	result := &batteryResult{
		Percentage: 100,
		Status:     "AC Connected",
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$battery = Get-CimInstance -ClassName Win32_Battery -ErrorAction SilentlyContinue
		if ($battery) {
			@{
				Percentage = $battery.EstimatedChargeRemaining
				Status = $battery.BatteryStatus
			} | ConvertTo-Json -Compress
		}
	`)

	output, err := cmd.Output()
	if err != nil {
		return result
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "null" {
		return result
	}

	var bat struct {
		Percentage int `json:"Percentage"`
		Status     int `json:"Status"`
	}
	if err := json.Unmarshal([]byte(outputStr), &bat); err == nil {
		result.Percentage = bat.Percentage
		result.Status = getBatteryStatusText(bat.Status)
	}

	return result
}

func (c *windowsCollector) collectNetwork() *networkResult {
	result := &networkResult{
		Interfaces: make([]interfaceResult, 0),
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		$ErrorActionPreference = 'SilentlyContinue'
		$adapters = Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.IPAddress -notlike '127.*' -and $_.IPAddress -notlike '169.254.*' } | ForEach-Object {
			$iface = Get-NetInterface -InterfaceIndex $_.InterfaceIndex -ErrorAction SilentlyContinue
			@{
				Name = if($iface) { $iface.InterfaceAlias } else { $_.InterfaceAlias }
				IP = $_.IPAddress
			}
		}
		@{
			Interfaces = $adapters
		} | ConvertTo-Json -Compress -Depth 2
	`)

	output, err := cmd.Output()
	if err == nil {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" && outputStr != "{}" {
			var netData struct {
				Interfaces []interfaceResult `json:"Interfaces"`
			}
			if err := json.Unmarshal([]byte(outputStr), &netData); err == nil {
				result.Interfaces = netData.Interfaces
				if len(result.Interfaces) > 0 {
					result.LocalIP = result.Interfaces[0].IP
				}
			}
		}
	}

	return result
}

func detectVendor(name string) string {
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, "nvidia") {
		return "NVIDIA"
	}
	if strings.Contains(nameLower, "amd") || strings.Contains(nameLower, "radeon") {
		return "AMD"
	}
	if strings.Contains(nameLower, "intel") {
		return "Intel"
	}
	return "Unknown"
}

func getBatteryStatusText(status int) string {
	switch status {
	case 1:
		return "Discharging"
	case 2:
		return "AC Connected"
	case 3:
		return "Fully Charged"
	case 4:
		return "Low"
	case 5:
		return "Critical"
	case 6:
		return "Charging"
	case 7:
		return "Charging High"
	case 8:
		return "Charging Low"
	case 9:
		return "Charging Critical"
	case 10:
		return "Undefined"
	case 11:
		return "Partially Charged"
	default:
		return "Unknown"
	}
}

func detectSystem() (any, error) {
	c := GetCollector()
	return convertToSystemInfo(c.SystemInfo), nil
}

func detectHardware() (any, error) {
	c := GetCollector()
	return convertToHardwareInfo(c.Hardware, c.Battery), nil
}

func detectNetwork() (any, error) {
	c := GetCollector()
	return convertToNetworkInfo(c.Network), nil
}

func convertToSystemInfo(ci *systemInfoResult) *SystemInfo {
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

func convertToHardwareInfo(hw *hardwareResult, bat *batteryResult) *HardwareInfo {
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

func convertToNetworkInfo(ni *networkResult) *NetworkInfo {
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
