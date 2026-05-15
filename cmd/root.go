package cmd

import (
	"github.com/gausszhou/hardfetch/internal/detect"
	"github.com/gausszhou/hardfetch/internal/display"
	"github.com/gausszhou/hardfetch/internal/logger"
	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "hardfetch",
	Short: "A system information fetching tool",
	Long:  "HardFetch is a command-line tool for fetching system information, similar to fastfetch/neofetch.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(debug)

		detectors := detect.GetCoreDetectors()
		result := detect.Detect(detectors...)
		display.PrintResult(result)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	rootCmd.Version = "0.1.0"
	rootCmd.SetVersionTemplate("hardfetch {{.Version}}\n")
}
