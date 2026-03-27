package network

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name        string
	MACAddress  string
	IPAddress   string
	IPv6Address string
}

// NetworkInfo represents network information
type NetworkInfo struct {
	Hostname   string
	LocalIP    string
	PublicIP   string
	Interfaces []NetworkInterface
}

// GetNetworkInfo collects network information
func GetNetworkInfo() (*NetworkInfo, error) {
	info := &NetworkInfo{}

	// Get hostname
	if hostname, err := getHostname(); err == nil {
		info.Hostname = hostname
	}

	// Get local IP
	if localIP, err := getLocalIP(); err == nil {
		info.LocalIP = localIP
	}

	// Get network interfaces
	interfaces, err := getNetworkInterfaces()
	if err == nil {
		info.Interfaces = interfaces
	}

	// Note: Public IP requires external API call
	// We'll leave it empty for now to avoid network calls

	return info, nil
}

func getHostname() (string, error) {
	// Use net package to get hostname
	return os.Hostname()
}

func getLocalIP() (string, error) {
	// Get first non-loopback IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no non-loopback IP address found")
}

func getNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []NetworkInterface
	for _, iface := range interfaces {
		// Skip loopback and inactive interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		ni := NetworkInterface{
			Name:       iface.Name,
			MACAddress: iface.HardwareAddr.String(),
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil && ni.IPAddress == "" {
						ni.IPAddress = ipnet.IP.String()
					} else if ipnet.IP.To4() == nil && ni.IPv6Address == "" {
						ni.IPv6Address = ipnet.IP.String()
					}
				}
			}
		}

		result = append(result, ni)
	}

	return result, nil
}

// FormatInterfaces formats network interfaces info
func (n *NetworkInfo) FormatInterfaces() string {
	if len(n.Interfaces) == 0 {
		return "No active interfaces"
	}

	var result []string
	for _, iface := range n.Interfaces {
		info := iface.Name
		if iface.IPAddress != "" {
			info += fmt.Sprintf(" (%s)", iface.IPAddress)
		}
		result = append(result, info)
	}
	return strings.Join(result, ", ")
}
