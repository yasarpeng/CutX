//go:build !windows

package gui

// Run is a no-op on non-Windows platforms.
func Run(version string) {}
