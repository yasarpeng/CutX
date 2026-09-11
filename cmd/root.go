package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cutx",
	Short: "CutX - Cross-platform offline file splitter & merger",
	Long: `CutX - Cross-platform offline file splitter & merger

A single binary that splits large files into smaller chunks and merges them back.
Works offline on macOS, Windows, and Linux. No runtime dependencies.
Data integrity via MD5 (default) or SHA-256 checksums.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// exitOnError prints error and exits with code 1.
func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ cutx ] ✗ Error: %s\n", err.Error())
		os.Exit(1)
	}
}
