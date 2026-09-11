package cmd

import (
	"github.com/pengyongshi/cutx/internal/engine"
	"github.com/spf13/cobra"
)

var (
	verifyQuiet   bool
	verifyVerbose bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify <manifest-file> [flags]",
	Short: "Verify chunk integrity without merging",
	Args:  cobra.ExactArgs(1),
	RunE:  runVerify,
}

func init() {
	verifyCmd.Flags().BoolVarP(&verifyQuiet, "quiet", "q", false, "Quiet mode")
	verifyCmd.Flags().BoolVarP(&verifyVerbose, "verbose", "v", false, "Verbose log mode")
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) error {
	rep := newCLIReporter(verifyQuiet)
	return engine.Verify(engine.VerifyOptions{
		ManifestPath: args[0],
	}, rep)
}
