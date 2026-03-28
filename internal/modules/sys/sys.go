//go:build windows

package sys

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v4/host"
	"golang.org/x/sys/windows"
)

type Info struct {
	OS       string
	Arch     string
	Kernel   string
	Hostname string
	Host     string
	Uptime   time.Duration
}

func Get() (*Info, error) {
	info := &Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: getHostname(),
	}

	info.Host = getWindowsHost()
	info.Kernel = getWindowsKernel()

	uptime, err := host.BootTime()
	if err == nil {
		info.Uptime = time.Duration(time.Now().Unix()-int64(uptime)) * time.Second
	}

	info.OS = "Windows"

	return info, nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}
	return hostname
}

func getWindowsHost() string {
	var hKey windows.Handle
	keyPath, _ := windows.UTF16PtrFromString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, keyPath, 0, windows.KEY_READ|windows.KEY_WOW64_64KEY, &hKey); err != nil {
		return ""
	}
	defer windows.RegCloseKey(hKey)

	productName := queryRegString(hKey, "ProductName")
	displayVersion := queryRegString(hKey, "DisplayVersion")

	if productName == "" {
		return ""
	}
	if displayVersion != "" {
		return productName + " " + displayVersion
	}
	return productName
}

func getWindowsKernel() string {
	var hKey windows.Handle
	keyPath, _ := windows.UTF16PtrFromString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err := windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, keyPath, 0, windows.KEY_READ|windows.KEY_WOW64_64KEY, &hKey); err != nil {
		return ""
	}
	defer windows.RegCloseKey(hKey)

	build := queryRegString(hKey, "CurrentBuild")
	ubr := queryRegString(hKey, "UBR")

	if build == "" {
		return ""
	}
	if ubr != "" {
		return build + "." + ubr
	}
	return build
}

func queryRegString(hKey windows.Handle, name string) string {
	namePtr, _ := windows.UTF16PtrFromString(name)
	var buf [32]byte
	var bufType uint32
	size := uint32(len(buf))
	err := windows.RegQueryValueEx(hKey, namePtr, nil, &bufType, &buf[0], &size)
	if err != nil {
		return ""
	}
	if bufType == windows.REG_DWORD {
		val := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
		return fmt.Sprintf("%d", val)
	}
	return windows.UTF16ToString((*[16]uint16)(unsafe.Pointer(&buf[0]))[:size/2])
}
