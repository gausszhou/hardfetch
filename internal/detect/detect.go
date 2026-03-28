package detect

import (
	"sync"
)

type Detector interface {
	Name() string
	Detect() (any, error)
}

type Result struct {
	System   *SystemInfo
	Hardware *HardwareInfo
	Network  *NetworkInfo
}

type coreDetector struct {
	name string
	fn   func() (any, error)
}

func (d *coreDetector) Name() string {
	return d.name
}

func (d *coreDetector) Detect() (any, error) {
	return d.fn()
}

var (
	result     *Result
	resultOnce sync.Once
)

func GetCoreDetectors() []Detector {
	return []Detector{
		&coreDetector{name: "system", fn: detectSystem},
		&coreDetector{name: "hardware", fn: detectHardware},
		&coreDetector{name: "network", fn: detectNetwork},
	}
}

func Detect(detectors ...Detector) *Result {
	resultOnce.Do(func() {
		result = &Result{}
		collectAll(detectors)
	})
	return result
}

func collectAll(detectors []Detector) {
	var wg sync.WaitGroup
	wg.Add(len(detectors))

	for _, d := range detectors {
		go func(detector Detector) {
			defer wg.Done()
			data, err := detector.Detect()
			if err != nil {
				return
			}
			switch detector.Name() {
			case "system":
				if sys, ok := data.(*SystemInfo); ok {
					result.System = sys
				}
			case "hardware":
				if hw, ok := data.(*HardwareInfo); ok {
					result.Hardware = hw
				}
			case "network":
				if net, ok := data.(*NetworkInfo); ok {
					result.Network = net
				}
			}
		}(d)
	}

	wg.Wait()
}
