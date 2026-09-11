package main

import (
	"os"
	"runtime"

	"github.com/pengyongshi/cutx/cmd"
	"github.com/pengyongshi/cutx/internal/gui"
)

var (
	version = "v1.0.0"
	commit  = "dev"
)

func main() {
	if len(os.Args) == 1 {
		if runtime.GOOS == "windows" {
			// Windows: launch native GUI window (like PuTTY)
			gui.Run(version)
			return
		}
		// macOS/Linux: interactive CLI menu
		cmd.RunInteractive()
		return
	}
	cmd.Execute()
}
