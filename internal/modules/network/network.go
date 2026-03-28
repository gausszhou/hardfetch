package network

import (
	"net"
	"os"
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
	info := &Info{
		Hostname:   "Unknown",
		LocalIP:    "",
		Interfaces: []Interface{},
	}

	hostname, _ := os.Hostname()
	info.Hostname = hostname

	ifaces, err := net.Interfaces()
	if err != nil {
		return info, nil
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip := parseIP(addr.String())
			if ip == "" || isLinkLocal(ip) {
				continue
			}

			info.Interfaces = append(info.Interfaces, Interface{
				Name:      iface.Name,
				IPAddress: ip,
			})
			if info.LocalIP == "" {
				info.LocalIP = ip
			}
		}
	}

	return info, nil
}

func parseIP(addr string) string {
	ip := net.ParseIP(addr)
	if ip == nil {
		return ""
	}
	return ip.String()
}

func isLinkLocal(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	ip4 := ip.To4()
	if ip4 != nil && ip4[0] == 169 && ip4[1] == 254 {
		return true
	}
	return strings.HasPrefix(ipStr, "fe80:")
}
