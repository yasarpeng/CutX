package main

import (
	"os"
	"runtime"

	"github.com/pengyongshi/cutx/cmd"
	"github.com/pengyongshi/cutx/internal/webui"
)

var (
	version = "v1.0.0"
	commit  = "dev"
)

func main() {
	// If no arguments provided:
	// - Windows: launch GUI (web server + browser)
	// - macOS/Linux: launch interactive CLI menu
	if len(os.Args) == 1 {
		if runtime.GOOS == "windows" {
			webui.Run(version, commit)
			return
		}
		cmd.RunInteractive()
		return
	}
	cmd.Execute()
}
