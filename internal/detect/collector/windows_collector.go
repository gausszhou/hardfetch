//go:build windows

package collector

import (
	"encoding/json"
	"os/exec"
	"strings"
	"sync"
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type systemInfo struct {
	wProcessorArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress uintptr
	lpMaximumApplicationAddress uintptr
	dwActiveProcessorMask       uintptr
	dwNumberOfProcessors        uint32
	dwProcessorType             uint32
	dwAllocationGranularity     uint32
	wProcessorLevel             uint16
	wProcessorRevision          uint16
}

var (
	collectorInstance *WindowsCollector
	collectorOnce     sync.Once
)

type WindowsCollector struct {
	mu         sync.RWMutex
	SystemInfo *SystemInfoResult
	Hardware   *HardwareResult
	Battery    *BatteryResult
	Network    *NetworkResult
}

type SystemInfoResult struct {
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

type GPUResult struct {
	Name          string `json:"Name"`
	Vendor        string `json:"Vendor"`
	VRAM          uint64 `json:"VRAM"`
	DriverVersion string `json:"DriverVersion"`
}

type HardwareResult struct {
	Memory      uint64       `json:"Memory"`
	MemoryUsed  uint64       `json:"MemoryUsed"`
	SwapTotal   uint64       `json:"SwapTotal"`
	SwapUsed    uint64       `json:"SwapUsed"`
	GPUs        []GPUResult  `json:"GPUs"`
	Disks       []DiskResult `json:"Disks"`
	MonitorName string       `json:"MonitorName"`
}

type DiskResult struct {
	Drive      string `json:"Drive"`
	Total      uint64 `json:"Total"`
	Used       uint64 `json:"Used"`
	Free       uint64 `json:"Free"`
	FileSystem string `json:"FileSystem"`
}

type BatteryResult struct {
	Percentage int    `json:"Percentage"`
	Status     string `json:"Status"`
}

type NetworkResult struct {
	LocalIP    string            `json:"LocalIP"`
	Interfaces []InterfaceResult `json:"Interfaces"`
}

type InterfaceResult struct {
	Name string `json:"Name"`
	IP   string `json:"IP"`
}

func GetCollector() *WindowsCollector {
	collectorOnce.Do(func() {
		collectorInstance = &WindowsCollector{}
		collectorInstance.collectAll()
	})
	return collectorInstance
}

func (c *WindowsCollector) collectAll() {
	var wg sync.WaitGroup

	wg.Add(4)
	go func() {
		defer wg.Done()
		c.SystemInfo = c.collectSystemInfo()
	}()
	go func() {
		defer wg.Done()
		c.Hardware = c.collectHardware()
	}()
	go func() {
		defer wg.Done()
		c.Battery = c.collectBattery()
	}()
	go func() {
		defer wg.Done()
		c.Network = c.collectNetwork()
	}()

	wg.Wait()
}

func (c *WindowsCollector) collectSystemInfo() *SystemInfoResult {
	result := &SystemInfoResult{}

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

func (c *WindowsCollector) collectHardware() *HardwareResult {
	result := &HardwareResult{
		GPUs:  make([]GPUResult, 0),
		Disks: make([]DiskResult, 0),
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
							gpu := GPUResult{}
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
							disk := DiskResult{}
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

func (c *WindowsCollector) collectBattery() *BatteryResult {
	result := &BatteryResult{
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

func (c *WindowsCollector) collectNetwork() *NetworkResult {
	result := &NetworkResult{
		Interfaces: make([]InterfaceResult, 0),
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
				Interfaces []InterfaceResult `json:"Interfaces"`
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
