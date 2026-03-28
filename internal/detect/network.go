package detect

import (
	"fmt"
	"strings"
)

type NetworkInterface struct {
	Name        string
	MACAddress  string
	IPAddress   string
	IPv6Address string
}

type NetworkInfo struct {
	Hostname   string
	LocalIP    string
	PublicIP   string
	Interfaces []NetworkInterface
}

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
