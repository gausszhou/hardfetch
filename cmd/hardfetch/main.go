package main

import (
	"fmt"
	"os"

	"github.com/gausszhou/hardfetch/internal/cli"
	"github.com/gausszhou/hardfetch/internal/detect"
	"github.com/gausszhou/hardfetch/internal/display"
	"github.com/gausszhou/hardfetch/internal/logger"
)

func main() {
	debugMode := false
	args := make([]string, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--debug", "-d":
			debugMode = true
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

	result := detect.Detect(detect.GetCoreDetectors()...)
	display.PrintResult(result)
}

func printVersion() {
	fmt.Printf("%s version %s\n", cli.Name, cli.Version)
	fmt.Printf("Author: %s\n", cli.Author)
	fmt.Printf("Repo: %s\n", cli.Repo)
}

func printHelp() {
	fmt.Printf("Usage: %s [options]\n", cli.Name)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("  -d, --debug    Enable debug logging for performance analysis")
}
