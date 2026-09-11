package engine

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pengyongshi/cutx/internal/chunk"
	"github.com/pengyongshi/cutx/internal/checksum"
	"github.com/pengyongshi/cutx/internal/manifest"
)

// Reporter abstracts progress and log output so both CLI and GUI can share the same engine.
type Reporter interface {
	Log(format string, args ...interface{})
	ProgressStart(total int64, label string)
	ProgressAdd(n int64)
	ProgressFinish()
	ProgressStop()
}

// SplitOptions configures a Split operation.
type SplitOptions struct {
	SourcePath string
	ChunkSize  string // e.g. "2G", "500M"
	OutputDir  string // empty = default subdirectory
	HashAlgo   string // "md5" or "sha256"
	ToolVer    string // tool version string for manifest
}

// MergeOptions configures a Merge operation.
type MergeOptions struct {
	ManifestPath string
	Mode         string // "quick" or "verify"
	OutputDir    string // empty = manifest directory
	Force        bool
}

// VerifyOptions configures a Verify operation.
type VerifyOptions struct {
	ManifestPath string
}

// Split performs the file splitting operation.
func Split(opts SplitOptions, rep Reporter) error {
	sourcePath := opts.SourcePath

	chunkSize, err := chunk.ParseSize(opts.ChunkSize)
	if err != nil {
		return err
	}

	if opts.HashAlgo != "md5" && opts.HashAlgo != "sha256" {
		return fmt.Errorf("--hash only accepts md5 or sha256")
	}

	baseName := chunk.BaseName(sourcePath)
	if strings.ContainsAny(baseName, "/\\") {
		return fmt.Errorf("filename contains path separator: %s", baseName)
	}

	outDir := opts.OutputDir
	if outDir == "" {
		outDir = filepath.Join(filepath.Dir(sourcePath), "cutx-"+chunk.StripExtensions(baseName))
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source file not found or unreadable: %s", sourcePath)
	}
	if info.IsDir() {
		return fmt.Errorf("source is a directory, not a file: %s", sourcePath)
	}
	fileSize := info.Size()

	realPath := sourcePath
	if linkTarget, err := filepath.EvalSymlinks(sourcePath); err == nil {
		realPath = linkTarget
	}

	chunkCount := chunk.CountChunks(fileSize, chunkSize)
	if chunkCount > chunk.MaxChunkCount {
		return fmt.Errorf("chunk count %d exceeds limit %d, please increase chunk size", chunkCount, chunk.MaxChunkCount)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory %s: %w", outDir, err)
	}

	rep.Log("Cleaning residual chunks...")
	if err := chunk.CleanResidual(outDir, baseName); err != nil {
		return fmt.Errorf("failed to clean residual files: %w", err)
	}

	if chunkSize > fileSize {
		rep.Log("⚠️  Chunk size (%s) > file size (%s), generating single chunk",
			chunk.FormatBytes(chunkSize), chunk.FormatBytes(fileSize))
	}

	src, err := os.Open(realPath)
	if err != nil {
		return fmt.Errorf("cannot open source file: %w", err)
	}
	defer src.Close()

	rep.ProgressStart(fileSize, fmt.Sprintf("Splitting: %s", baseName))
	startTime := time.Now()

	m := manifest.New(opts.ToolVer, baseName, fileSize, info.ModTime(), chunkSize, chunkCount, opts.HashAlgo)
	overallHasher, err := checksum.New(opts.HashAlgo)
	if err != nil {
		return err
	}

	buf := make([]byte, 4*1024*1024)
	chunkIdx := 1
	var currentChunkFile *os.File
	var currentChunkHasher hash.Hash
	var currentChunkWritten int64

	openNewChunk := func() error {
		chunkName := chunk.ChunkFileName(baseName, chunkIdx)
		chunkPath := filepath.Join(outDir, chunkName)
		f, err := os.Create(chunkPath)
		if err != nil {
			return fmt.Errorf("cannot create chunk file %s: %w", chunkPath, err)
		}
		currentChunkFile = f
		currentChunkHasher, err = checksum.New(opts.HashAlgo)
		if err != nil {
			f.Close()
			return err
		}
		currentChunkWritten = 0
		return nil
	}

	closeCurrentChunk := func() {
		if currentChunkFile != nil {
			currentChunkFile.Close()
			chunkHash := hex.EncodeToString(currentChunkHasher.Sum(nil))
			m.Chunks = append(m.Chunks, manifest.Chunk{
				Index:    chunkIdx,
				Filename: chunk.ChunkFileName(baseName, chunkIdx),
				Size:     currentChunkWritten,
				Hash:     chunkHash,
			})
		}
	}

	if err := openNewChunk(); err != nil {
		return err
	}

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			data := buf[:n]
			for len(data) > 0 {
				remaining := chunkSize - currentChunkWritten
				writeLen := int64(len(data))
				if writeLen > remaining {
					writeLen = remaining
				}
				if _, err := currentChunkFile.Write(data[:writeLen]); err != nil {
					currentChunkFile.Close()
					return fmt.Errorf("write error on chunk %d: %w", chunkIdx, err)
				}
				currentChunkHasher.Write(data[:writeLen])
				overallHasher.Write(data[:writeLen])
				currentChunkWritten += writeLen
				rep.ProgressAdd(writeLen)
				data = data[writeLen:]
				if currentChunkWritten >= chunkSize && chunkIdx < chunkCount {
					closeCurrentChunk()
					chunkIdx++
					if err := openNewChunk(); err != nil {
						return err
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		} else if readErr != nil {
			if currentChunkFile != nil {
				currentChunkFile.Close()
			}
			return fmt.Errorf("read error: %w", readErr)
		}
	}

	closeCurrentChunk()
	rep.ProgressFinish()
	elapsed := time.Since(startTime)

	m.OverallHash = hex.EncodeToString(overallHasher.Sum(nil))

	manifestPath := filepath.Join(outDir, chunk.ManifestFileName(baseName))
	if err := m.Save(manifestPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	rep.Log("✓ Done! %d parts + manifest created in %s", chunkCount, formatDuration(elapsed))
	rep.Log("Output: %s", outDir)
	rep.Log("Manifest: %s", manifestPath)
	rep.Log("Tip: Verify with → cutx verify %s", manifestPath)
	rep.Log("Tip: Merge with → cutx merge %s", manifestPath)

	return nil
}

// Merge performs the file merging operation.
func Merge(opts MergeOptions, rep Reporter) error {
	manifestPath := opts.ManifestPath

	if opts.Mode != "quick" && opts.Mode != "verify" {
		return fmt.Errorf("--mode only accepts quick or verify, got: %s", opts.Mode)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	manifestDir := filepath.Dir(manifestPath)
	outDir := opts.OutputDir
	if outDir == "" {
		outDir = manifestDir
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	outputPath := filepath.Join(outDir, m.Source.Filename)
	if _, err := os.Stat(outputPath); err == nil && !opts.Force {
		return fmt.Errorf("output file already exists: %s (use --force to overwrite)", outputPath)
	}

	rep.Log("Pre-check: checking %d parts...", len(m.Chunks))
	missingParts, sizeMismatches := precheckChunks(manifestDir, m)
	if len(missingParts) > 0 || len(sizeMismatches) > 0 {
		for _, mp := range missingParts {
			rep.Log("✗ Missing: %s (part %d)", mp.Filename, mp.Index)
		}
		for _, sm := range sizeMismatches {
			rep.Log("✗ Size mismatch: %s (expected %d, got %d)", sm.Filename, sm.ExpectedSize, sm.ActualSize)
		}
		return fmt.Errorf("pre-check failed: %d missing, %d size mismatch", len(missingParts), len(sizeMismatches))
	}

	freeSpace, err := chunk.DiskFree(outDir)
	if err == nil && freeSpace < m.Source.Size {
		return fmt.Errorf("insufficient disk space: need %s, available %s",
			chunk.FormatBytes(m.Source.Size), chunk.FormatBytes(freeSpace))
	}

	rep.Log("Pre-check: ✓ All parts found, disk space sufficient")

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	startTime := time.Now()
	rep.ProgressStart(m.Source.Size, fmt.Sprintf("Merging: %s", m.Source.Filename))

	var overallHasher hash.Hash
	if opts.Mode == "verify" {
		overallHasher, err = checksum.New(m.Split.HashAlgorithm)
		if err != nil {
			return err
		}
	}

	buf := make([]byte, 4*1024*1024)
	mergeError := false

	for _, c := range m.Chunks {
		chunkPath := filepath.Join(manifestDir, c.Filename)
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("cannot open chunk %s: %w", c.Filename, err)
		}

		var chunkHasher hash.Hash
		if opts.Mode == "verify" {
			chunkHasher, err = checksum.New(m.Split.HashAlgorithm)
			if err != nil {
				chunkFile.Close()
				return err
			}
		}

		for {
			n, readErr := chunkFile.Read(buf)
			if n > 0 {
				if _, err := outFile.Write(buf[:n]); err != nil {
					chunkFile.Close()
					return fmt.Errorf("write error: %w", err)
				}
				if opts.Mode == "verify" {
					chunkHasher.Write(buf[:n])
					overallHasher.Write(buf[:n])
				}
				rep.ProgressAdd(int64(n))
			}
			if readErr == io.EOF {
				break
			} else if readErr != nil {
				chunkFile.Close()
				return fmt.Errorf("read error on %s: %w", c.Filename, readErr)
			}
		}
		chunkFile.Close()

		if opts.Mode == "verify" {
			actualHash := fmt.Sprintf("%x", chunkHasher.Sum(nil))
			if actualHash != c.Hash {
				outFile.Close()
				os.Remove(outputPath)
				rep.Log("✗ Hash mismatch on %s (part %d): expected %s, got %s",
					c.Filename, c.Index, c.Hash, actualHash)
				mergeError = true
				break
			}
		}
	}

	rep.ProgressFinish()

	if mergeError {
		return fmt.Errorf("merge aborted: chunk hash verification failed")
	}

	elapsed := time.Since(startTime)

	if opts.Mode == "verify" && m.OverallHash != "" {
		actualOverall := fmt.Sprintf("%x", overallHasher.Sum(nil))
		if actualOverall != m.OverallHash {
			rep.Log("⚠️  Overall hash mismatch: expected %s, got %s", m.OverallHash, actualOverall)
			rep.Log("⚠️  File may be incomplete. This should not happen if all chunk hashes matched.")
			rep.Log("✓ Done! File restored in %s (with warning)", formatDuration(elapsed))
			rep.Log("Output: %s (%s)", outputPath, chunk.FormatBytes(m.Source.Size))
			return nil
		}
		rep.Log("Overall hash: ✓ Verified")
	}

	rep.Log("✓ Done! File restored in %s", formatDuration(elapsed))
	rep.Log("Output: %s (%s)", outputPath, chunk.FormatBytes(m.Source.Size))

	return nil
}

// Verify performs chunk integrity verification without merging.
func Verify(opts VerifyOptions, rep Reporter) error {
	manifestPath := opts.ManifestPath

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	manifestDir := filepath.Dir(manifestPath)

	rep.Log("Verifying: %s (%d parts, %s)", m.Source.Filename, len(m.Chunks), chunk.FormatBytes(m.Source.Size))
	rep.Log("Manifest: ✓ Valid JSON, all required fields present")
	rep.Log("Algorithm: %s", m.Split.HashAlgorithm)

	missingParts, sizeMismatches := precheckChunks(manifestDir, m)
	if len(missingParts) > 0 || len(sizeMismatches) > 0 {
		for _, mp := range missingParts {
			rep.Log("✗ Missing: %s (part %d)", mp.Filename, mp.Index)
		}
		for _, sm := range sizeMismatches {
			rep.Log("✗ Size mismatch: %s (expected %d, got %d)", sm.Filename, sm.ExpectedSize, sm.ActualSize)
		}
		return fmt.Errorf("pre-check failed: %d missing, %d size mismatch", len(missingParts), len(sizeMismatches))
	}

	var totalSize int64
	for _, c := range m.Chunks {
		totalSize += c.Size
	}

	startTime := time.Now()
	rep.ProgressStart(totalSize, "Verifying")
	buf := make([]byte, 4*1024*1024)

	failedCount := 0
	verifiedCount := 0

	for _, c := range m.Chunks {
		chunkPath := filepath.Join(manifestDir, c.Filename)
		f, err := os.Open(chunkPath)
		if err != nil {
			rep.Log("✗ Cannot open: %s", c.Filename)
			failedCount++
			continue
		}

		h, err := checksum.New(m.Split.HashAlgorithm)
		if err != nil {
			f.Close()
			return err
		}

		for {
			n, readErr := f.Read(buf)
			if n > 0 {
				h.Write(buf[:n])
				rep.ProgressAdd(int64(n))
			}
			if readErr == io.EOF {
				break
			} else if readErr != nil {
				f.Close()
				return fmt.Errorf("read error on %s: %w", c.Filename, readErr)
			}
		}
		f.Close()

		actualHash := fmt.Sprintf("%x", h.Sum(nil))
		if actualHash != c.Hash {
			rep.Log("✗ Hash mismatch: %s (part %d): expected %s, got %s",
				c.Filename, c.Index, c.Hash, actualHash)
			failedCount++
		} else {
			verifiedCount++
		}
	}

	rep.ProgressFinish()
	elapsed := time.Since(startTime)

	if failedCount > 0 {
		rep.Log("✗ Verification failed: %d/%d parts failed", failedCount, len(m.Chunks))
		return fmt.Errorf("verification failed: %d/%d parts failed", failedCount, len(m.Chunks))
	}

	rep.Log("✓ All %d parts verified successfully! (%s)", verifiedCount, formatDuration(elapsed))
	return nil
}

// --- shared helpers ---

type missingPart struct {
	Index    int
	Filename string
}

type sizeMismatch struct {
	Index        int
	Filename     string
	ExpectedSize int64
	ActualSize   int64
}

func precheckChunks(dir string, m *manifest.Manifest) ([]missingPart, []sizeMismatch) {
	var missing []missingPart
	var mismatches []sizeMismatch
	for _, c := range m.Chunks {
		chunkPath := filepath.Join(dir, c.Filename)
		info, err := os.Stat(chunkPath)
		if err != nil {
			missing = append(missing, missingPart{Index: c.Index, Filename: c.Filename})
			continue
		}
		if info.Size() != c.Size {
			mismatches = append(mismatches, sizeMismatch{
				Index:        c.Index,
				Filename:     c.Filename,
				ExpectedSize: c.Size,
				ActualSize:   info.Size(),
			})
		}
	}
	return missing, mismatches
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", m, s)
}
