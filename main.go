package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/gausszhou/hardfetch/internal/detect"
	"github.com/gausszhou/hardfetch/internal/display"
	"github.com/gausszhou/hardfetch/internal/info"
	"github.com/gausszhou/hardfetch/internal/logger"
)

func main() {
	debugMode := false
	pprofMode := false
	args := make([]string, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--debug", "-d":
			debugMode = true
		case "--pprof", "-p":
			pprofMode = true
		case "--version", "-v":
			printVersion()
			return
		case "--help", "-h":
			printHelp()
			return
		default:
			args = append(args, arg)
		}
	}

	logger.Init(debugMode)

	var cpuProfile *os.File
	if pprofMode {
		f, err := os.Create("hardfetch_cpu.pprof")
		if err != nil {
			fmt.Printf("failed to create CPU profile: %v\n", err)
		} else {
			cpuProfile = f
			pprof.StartCPUProfile(cpuProfile)
			fmt.Printf("[pprof] CPU profiling started\n")
		}
	}

	result := detect.Detect(detect.GetCoreDetectors()...)
	display.PrintResult(result)

	if pprofMode {
		pprof.StopCPUProfile()
		if cpuProfile != nil {
			cpuProfile.Close()
			fmt.Printf("[pprof] CPU profile saved to hardfetch_cpu.pprof\n")
		}

		memProfileFile, err := os.Create("hardfetch_mem.pprof")
		if err != nil {
			fmt.Printf("failed to create memory profile: %v\n", err)
		} else {
			pprof.WriteHeapProfile(memProfileFile)
			memProfileFile.Close()
			fmt.Printf("[pprof] Memory profile saved to hardfetch_mem.pprof\n")
		}
	}
}

func printVersion() {
	fmt.Printf("%s version %s\n", info.Name, info.Version)
	fmt.Printf("Author: %s\n", info.Author)
	fmt.Printf("Repo: %s\n", info.Repo)
}

func printHelp() {
	fmt.Printf("Usage: %s [options]\n", info.Name)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("  -d, --debug    Enable debug logging")
	fmt.Println("  -p, --pprof    Generate pprof profile files")
}
