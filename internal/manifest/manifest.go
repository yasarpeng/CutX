package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ManifestVersion is the current manifest format version.
const ManifestVersion = 1

// Manifest is the JSON manifest file structure.
type Manifest struct {
	ManifestVersion int       `json:"manifest_version"`
	ToolVersion     string    `json:"tool_version"`
	CreatedAt      string    `json:"created_at"`
	Source         Source    `json:"source"`
	Split          Split     `json:"split"`
	Chunks         []Chunk   `json:"chunks"`
	OverallHash    string    `json:"overall_hash"`
}

type Source struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	ModTime  string `json:"mod_time"`
}

type Split struct {
	ChunkSize     int64  `json:"chunk_size"`
	ChunkCount    int    `json:"chunk_count"`
	HashAlgorithm string `json:"hash_algorithm"`
}

type Chunk struct {
	Index    int    `json:"index"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
}

// New creates a Manifest with sensible defaults.
func New(toolVersion, filename string, fileSize int64, modTime time.Time, chunkSize int64, chunkCount int, hashAlgo string) *Manifest {
	return &Manifest{
		ManifestVersion: ManifestVersion,
		ToolVersion:     toolVersion,
		CreatedAt:       time.Now().Format(time.RFC3339),
		Source: Source{
			Filename: filename,
			Size:     fileSize,
			ModTime:  modTime.Format(time.RFC3339),
		},
		Split: Split{
			ChunkSize:     chunkSize,
			ChunkCount:    chunkCount,
			HashAlgorithm: hashAlgo,
		},
		Chunks: []Chunk{},
	}
}

// Save writes the manifest to a JSON file.
func (m *Manifest) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}
	return nil
}

// Load reads and parses a manifest JSON file.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}
	return &m, nil
}

// Validate checks the manifest for required fields and consistency.
func (m *Manifest) Validate() error {
	if m.ManifestVersion <= 0 {
		return fmt.Errorf("invalid manifest_version: %d", m.ManifestVersion)
	}
	if m.Source.Filename == "" {
		return fmt.Errorf("missing required field: source.filename")
	}
	if m.Source.Size <= 0 {
		return fmt.Errorf("missing or invalid field: source.size")
	}
	if m.Split.ChunkSize <= 0 {
		return fmt.Errorf("missing or invalid field: split.chunk_size")
	}
	if m.Split.ChunkCount <= 0 {
		return fmt.Errorf("missing or invalid field: split.chunk_count")
	}
	if m.Split.HashAlgorithm == "" {
		return fmt.Errorf("missing required field: split.hash_algorithm")
	}
	if m.Split.HashAlgorithm != "md5" && m.Split.HashAlgorithm != "sha256" {
		return fmt.Errorf("unsupported hash algorithm: %s", m.Split.HashAlgorithm)
	}
	if len(m.Chunks) == 0 {
		return fmt.Errorf("manifest contains no chunks")
	}
	if len(m.Chunks) != m.Split.ChunkCount {
		return fmt.Errorf("chunk count mismatch: manifest says %d, chunks array has %d", m.Split.ChunkCount, len(m.Chunks))
	}
	for i, c := range m.Chunks {
		if c.Index != i+1 {
			return fmt.Errorf("chunk index mismatch at position %d: expected %d, got %d", i, i+1, c.Index)
		}
		if c.Filename == "" {
			return fmt.Errorf("chunk %d has empty filename", c.Index)
		}
		if c.Hash == "" {
			return fmt.Errorf("chunk %d has empty hash", c.Index)
		}
	}
	return nil
}
