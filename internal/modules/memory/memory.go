package memory

import (
	"github.com/shirou/gopsutil/v4/mem"
)

type Info struct {
	Total     uint64
	Used      uint64
	Available uint64
	Free      uint64
}

type SwapInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Available uint64
}

func Get() (*Info, *SwapInfo, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return &Info{}, &SwapInfo{}, err
	}

	info := &Info{
		Total:     v.Total,
		Used:      v.Used,
		Available: v.Available,
		Free:      v.Free,
	}

	swap, err := mem.SwapMemory()
	swapInfo := &SwapInfo{}
	if err == nil {
		swapInfo.Total = swap.Total
		swapInfo.Used = swap.Used
		swapInfo.Free = swap.Free
		swapInfo.Available = swap.Total - swap.Used
	}

	return info, swapInfo, nil
}
