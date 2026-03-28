package main

import (
	"fmt"
	"os"

	"github.com/gausszhou/hardfetch/internal/cli"
	"github.com/gausszhou/hardfetch/internal/detect"
	"github.com/gausszhou/hardfetch/internal/display"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			printVersion()
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	result := detect.Detect(detect.GetCoreDetectors()...)
	display.PrintResult(result)
}

func printVersion() {
	fmt.Printf("%s version %s\n", cli.Name, cli.Version)
	fmt.Printf("Author: %s\n", cli.Author)
	fmt.Printf("Website: %s\n", cli.Website)
}

func printHelp() {
	fmt.Printf("Usage: %s [options]\n", cli.Name)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
}
