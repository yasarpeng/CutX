package chunk

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const MaxChunkCount = 9999

// ParseSize parses a size string like "2G", "500M", "100K" into bytes.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, fmt.Errorf("size cannot be empty")
	}
	unit := strings.ToUpper(string(s[len(s)-1]))
	numStr := s[:len(s)-1]
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %s", numStr)
	}
	if num <= 0 {
		return 0, fmt.Errorf("size must be a positive number, got %d", num)
	}
	var multiplier int64
	switch unit {
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "M":
		multiplier = 1024 * 1024
	case "K":
		multiplier = 1024
	default:
		return 0, fmt.Errorf("invalid size unit: %s (use G, M, or K)", unit)
	}
	return num * multiplier, nil
}

// FormatBytes formats a byte count into human-readable string.
func FormatBytes(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.1f TB", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// ChunkFileName generates the chunk file name: {baseName}.part{4-digit index}
func ChunkFileName(baseName string, index int) string {
	return fmt.Sprintf("%s.part%04d", baseName, index)
}

// ManifestFileName generates the manifest file name: {baseName}.manifest.json
func ManifestFileName(baseName string) string {
	return fmt.Sprintf("%s.manifest.json", baseName)
}

// BaseName extracts the file name (without directory) from a path.
func BaseName(path string) string {
	return filepath.Base(path)
}

// StripExtensions removes all extensions from a filename, not just the last one.
// e.g. "archive.tar.gz" -> "archive", "ubuntu-server.img" -> "ubuntu-server".
// Dotfiles like ".bashrc" are left as-is (no extension to strip).
// If stripping would result in an empty string, the original name is returned.
func StripExtensions(name string) string {
	result := name
	for {
		ext := filepath.Ext(result)
		if ext == "" || ext == "." {
			break
		}
		result = strings.TrimSuffix(result, ext)
	}
	if result == "" {
		return name
	}
	return result
}

// CountChunks calculates how many chunks a file of totalSize bytes will produce.
func CountChunks(totalSize int64, chunkSize int64) int {
	if totalSize <= 0 {
		return 1
	}
	return int((totalSize + chunkSize - 1) / chunkSize)
}

// CleanResidual removes files matching {baseName}.part* and {baseName}.manifest.json in dir.
func CleanResidual(dir, baseName string) error {
	pattern := regexp.MustCompile(regexp.QuoteMeta(baseName) + `\.part\d+$`)
	manifestName := ManifestFileName(baseName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == manifestName {
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
			continue
		}
		if pattern.MatchString(name) {
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	}
	return nil
}

// DiskFree returns available disk space in bytes for the given path.
func DiskFree(path string) (int64, error) {
	// Resolve to actual directory
	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}
	// Make sure the directory exists
	for {
		info, err := os.Stat(abs)
		if err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return getDiskFree(abs)
}
