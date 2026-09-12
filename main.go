package main

import (
	"os"

	"github.com/pengyongshi/cutx/cmd"
)

var (
	version = "v1.0.0"
	commit  = "dev"
)

func main() {
	if len(os.Args) == 1 {
		// No arguments: launch interactive CLI menu
		// (Windows users should use the Wails GUI build: CutX.exe from gui/)
		cmd.RunInteractive()
		return
	}
	cmd.Execute()
}
