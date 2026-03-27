//go:build windows

package system

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

func getHostname() (string, error) {
	return os.Hostname()
}

func getKernelVersion() (string, error) {
	// On Windows, we can get the OS version
	// For now, return a simple identifier
	return "Windows", nil
}

func getUptime() (time.Duration, error) {
	// Get system uptime using GetTickCount64
	ret, _, err := procGetTickCount64.Call()
	if ret == 0 {
		return 0, fmt.Errorf("failed to get uptime: %v", err)
	}

	// GetTickCount64 returns milliseconds since system start
	uptimeMs := uint64(ret)
	return time.Duration(uptimeMs) * time.Millisecond, nil
}

// GetWindowsVersion gets detailed Windows version information
func GetWindowsVersion() (string, error) {
	type OSVersionInfoEx struct {
		OSVersionInfoSize uint32
		MajorVersion      uint32
		MinorVersion      uint32
		BuildNumber       uint32
		PlatformId        uint32
		CSDVersion        [128]uint16
		ServicePackMajor  uint16
		ServicePackMinor  uint16
		SuiteMask         uint16
		ProductType       byte
		Reserved          byte
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetVersionExW := kernel32.NewProc("GetVersionExW")

	var osInfo OSVersionInfoEx
	osInfo.OSVersionInfoSize = uint32(unsafe.Sizeof(osInfo))

	ret, _, err := procGetVersionExW.Call(uintptr(unsafe.Pointer(&osInfo)))
	if ret == 0 {
		return "Windows", fmt.Errorf("failed to get Windows version: %v", err)
	}

	return fmt.Sprintf("Windows %d.%d.%d", osInfo.MajorVersion, osInfo.MinorVersion, osInfo.BuildNumber), nil
}
