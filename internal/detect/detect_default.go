//go:build !windows

package detect

func detectSystem() (any, error) {
	return &SystemInfo{}, nil
}

func detectHardware() (any, error) {
	return &HardwareInfo{}, nil
}

func detectNetwork() (any, error) {
	return &NetworkInfo{}, nil
}
