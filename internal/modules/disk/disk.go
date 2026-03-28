package disk

import (
	"github.com/shirou/gopsutil/v4/disk"
)

type Info struct {
	Drive      string
	Total      uint64
	Used       uint64
	Free       uint64
	FileSystem string
}

func Get() ([]*Info, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	result := make([]*Info, 0)
	for _, p := range parts {
		if p.Mountpoint == "" {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		info := &Info{
			Drive:      p.Device,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			FileSystem: p.Fstype,
		}
		result = append(result, info)
	}

	return result, nil
}
