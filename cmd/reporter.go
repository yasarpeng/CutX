package cmd

import (
	"fmt"
	"os"

	"github.com/pengyongshi/cutx/internal/engine"
	"github.com/pengyongshi/cutx/internal/progress"
)

// cliReporter implements engine.Reporter for terminal output.
type cliReporter struct {
	pb     *progress.Bar
	quiet  bool
}

func newCLIReporter(quiet bool) *cliReporter {
	return &cliReporter{quiet: quiet}
}

func (r *cliReporter) Log(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ cutx ] "+format+"\n", args...)
}

func (r *cliReporter) ProgressStart(total int64, label string) {
	r.pb = progress.New(total, label, r.quiet)
}

func (r *cliReporter) ProgressAdd(n int64) {
	if r.pb != nil {
		r.pb.Add(n)
	}
}

func (r *cliReporter) ProgressFinish() {
	if r.pb != nil {
		r.pb.Finish()
	}
}

func (r *cliReporter) ProgressStop() {
	if r.pb != nil {
		r.pb.Stop()
	}
}

// compile-time interface check
var _ engine.Reporter = (*cliReporter)(nil)
