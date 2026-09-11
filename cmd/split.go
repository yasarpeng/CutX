package cmd

import (
	"github.com/pengyongshi/cutx/internal/engine"
	"github.com/spf13/cobra"
)

var (
	splitSize    string
	splitOutput  string
	splitHash    string
	splitQuiet   bool
	splitVerbose bool
)

var splitCmd = &cobra.Command{
	Use:   "split <source-file> [flags]",
	Short: "Split a large file into smaller chunks",
	Args:  cobra.ExactArgs(1),
	RunE:  runSplit,
}

func init() {
	splitCmd.Flags().StringVarP(&splitSize, "size", "s", "2G", "Chunk size (e.g. 2G, 500M, 100K)")
	splitCmd.Flags().StringVarP(&splitOutput, "output", "o", "", "Output directory (default: cutx-<filename>/ subdirectory next to source)")
	splitCmd.Flags().StringVar(&splitHash, "hash", "md5", "Hash algorithm: md5 or sha256")
	splitCmd.Flags().BoolVarP(&splitQuiet, "quiet", "q", false, "Quiet mode, no progress display")
	splitCmd.Flags().BoolVarP(&splitVerbose, "verbose", "v", false, "Verbose log mode")
	rootCmd.AddCommand(splitCmd)
}

func runSplit(cmd *cobra.Command, args []string) error {
	rep := newCLIReporter(splitQuiet)
	return engine.Split(engine.SplitOptions{
		SourcePath: args[0],
		ChunkSize:  splitSize,
		OutputDir:  splitOutput,
		HashAlgo:  splitHash,
		ToolVer:   version + " (" + commit + ")",
	}, rep)
}
