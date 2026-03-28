package network

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
)

type Interface struct {
	Name       string
	IPAddress  string
	MACAddress string
}

type Info struct {
	Hostname   string
	LocalIP    string
	PublicIP   string
	Interfaces []Interface
}

func Get() (*Info, error) {
	switch runtime.GOOS {
	case "windows":
		return getNetworkInfoWindows()
	case "darwin":
		return getNetworkInfoDarwin()
	case "linux":
		return getNetworkInfoLinux()
	default:
		return &Info{Hostname: "Unknown", LocalIP: "", Interfaces: []Interface{}}, nil
	}
}

func getNetworkInfoWindows() (*Info, error) {
	cmd := exec.Command("hostname")
	output, _ := cmd.Output()
	hostname := strings.TrimSpace(string(output))

	cmd = exec.Command("powershell", "-NoProfile", "-Command", `$ErrorActionPreference='SilentlyContinue';Get-NetIPAddress -AddressFamily IPv4|Where-Object{$_.IPAddress -notlike'127.*' -and $_.IPAddress -notlike'169.254.*'}|Select-Object InterfaceAlias,IPAddress|ConvertTo-Json -Compress`)
	output, err := cmd.Output()
	if err != nil {
		return &Info{Hostname: hostname, LocalIP: "", Interfaces: []Interface{}}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "{}" {
		return &Info{Hostname: hostname, LocalIP: "", Interfaces: []Interface{}}, nil
	}

	type adapterData struct {
		Name string `json:"InterfaceAlias"`
		IP   string `json:"IPAddress"`
	}

	var adapters []adapterData
	if strings.HasPrefix(outputStr, "[") {
		if err := json.Unmarshal([]byte(outputStr), &adapters); err != nil {
			return &Info{Hostname: hostname, LocalIP: "", Interfaces: []Interface{}}, nil
		}
	} else {
		var adapter adapterData
		if err := json.Unmarshal([]byte(outputStr), &adapter); err != nil {
			return &Info{Hostname: hostname, LocalIP: "", Interfaces: []Interface{}}, nil
		}
		adapters = []adapterData{adapter}
	}

	interfaces := make([]Interface, 0, len(adapters))
	var localIP string
	for _, iface := range adapters {
		interfaces = append(interfaces, Interface{
			Name:      iface.Name,
			IPAddress: iface.IP,
		})
		if localIP == "" && iface.IP != "" {
			localIP = iface.IP
		}
	}

	return &Info{
		Hostname:   hostname,
		LocalIP:    localIP,
		Interfaces: interfaces,
	}, nil
}

func getNetworkInfoDarwin() (*Info, error) {
	cmd := exec.Command("hostname")
	output, _ := cmd.Output()
	hostname := strings.TrimSpace(string(output))

	cmd = exec.Command("ipconfig", "getifaddr", "en0")
	output, _ = cmd.Output()
	localIP := strings.TrimSpace(string(output))

	cmd = exec.Command("networksetup", "-listallhardwareports")
	output, _ = cmd.Output()

	interfaces := make([]Interface, 0)
	lines := strings.Split(string(output), "\n")
	currentName := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "Hardware Port:") {
			currentName = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.HasPrefix(line, "Device:") && currentName != "" {
			dev := strings.TrimSpace(strings.Split(line, ":")[1])
			interfaces = append(interfaces, Interface{
				Name:      currentName,
				IPAddress: dev,
			})
		}
	}

	return &Info{
		Hostname:   hostname,
		LocalIP:    localIP,
		Interfaces: interfaces,
	}, nil
}

func getNetworkInfoLinux() (*Info, error) {
	cmd := exec.Command("hostname")
	output, _ := cmd.Output()
	hostname := strings.TrimSpace(string(output))

	cmd = exec.Command("hostname", "-I")
	output, _ = cmd.Output()
	ips := strings.Fields(strings.TrimSpace(string(output)))

	interfaces := make([]Interface, 0)
	for i, ip := range ips {
		ifname := "eth" + string(rune('0'+i))
		if len(ips) > 1 {
			ifname = "wlan" + string(rune('0'+i))
		}
		interfaces = append(interfaces, Interface{
			Name:      ifname,
			IPAddress: ip,
		})
	}

	localIP := ""
	if len(ips) > 0 {
		localIP = ips[0]
	}

	return &Info{
		Hostname:   hostname,
		LocalIP:    localIP,
		Interfaces: interfaces,
	}, nil
}
