//go:build !windows

package main

import "fmt"

func main() {
	fmt.Println("cutx-gui is only available on Windows. Use the CLI version (cutx) on this platform.")
}
