package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cutx %s\n", version)
		fmt.Printf("  built with: %s\n", runtime.Version())
		fmt.Printf("  platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  commit:     %s\n", commit)
		fmt.Printf("  built at:   %s\n", buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
