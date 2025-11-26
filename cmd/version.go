package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("netgaze v%s\n", version)
		fmt.Printf("Built: %s\n", buildDate)
		fmt.Printf("Go version: %s\n", runtime.Version())
	},
}
