package cmd

import (
	"fmt"
	"os"

	"github.com/pengyongshi/cutx/internal/chunk"
	"github.com/pengyongshi/cutx/internal/engine"
	"github.com/spf13/cobra"
)

var (
	mergeMode    string
	mergeOutput  string
	mergeForce   bool
	mergeQuiet   bool
	mergeVerbose bool
)

var mergeCmd = &cobra.Command{
	Use:   "merge <manifest-file> [flags]",
	Short: "Merge chunks back to the original file",
	Args:  cobra.ExactArgs(1),
	RunE:  runMerge,
}

func init() {
	mergeCmd.Flags().StringVar(&mergeMode, "mode", "quick", "Merge mode: quick (no verify) or verify (hash check while merging)")
	mergeCmd.Flags().StringVarP(&mergeOutput, "output", "o", "", "Output directory (default: same as manifest)")
	mergeCmd.Flags().BoolVar(&mergeForce, "force", false, "Force overwrite existing output file")
	mergeCmd.Flags().BoolVarP(&mergeQuiet, "quiet", "q", false, "Quiet mode")
	mergeCmd.Flags().BoolVarP(&mergeVerbose, "verbose", "v", false, "Verbose log mode")
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
	rep := newCLIReporter(mergeQuiet)
	err := engine.Merge(engine.MergeOptions{
		ManifestPath: args[0],
		Mode:         mergeMode,
		OutputDir:    mergeOutput,
		Force:        mergeForce,
	}, rep)
	if err != nil {
		return err
	}
	return nil
}

// formatBytes is used by merge summary output in the CLI wrapper if needed.
func formatBytes(b int64) string {
	return chunk.FormatBytes(b)
}

// exitCode2 prints a warning and exits with code 2 (used for overall hash mismatch).
func exitCode2(msg string) {
	fmt.Fprintf(os.Stderr, "[ cutx ] %s\n", msg)
	os.Exit(2)
}
