package main

import (
	"flag"
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

	flag.Parse()

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
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hardfetch                    # Show default system information")
	fmt.Println("  hardfetch --modules system,cpu,memory  # Show specific modules")
	fmt.Println("  hardfetch --all              # Show all available modules")
	fmt.Println("  hardfetch --gen-config       # Generate default configuration file")
}
