package checksum

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

// New creates a new hash.Hash for the given algorithm.
func New(algorithm string) (hash.Hash, error) {
	switch algorithm {
	case "md5":
		return md5.New(), nil
	case "sha256":
		return sha256.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s (use md5 or sha256)", algorithm)
	}
}

// Sum returns the hex-encoded hash of the given data.
func Sum(algorithm string, data []byte) (string, error) {
	h, err := New(algorithm)
	if err != nil {
		return "", err
	}
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SumReader reads all data from r and returns the hex-encoded hash.
func SumReader(algorithm string, r io.Reader) (string, error) {
	h, err := New(algorithm)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashLength returns the expected hex string length for the algorithm.
func HashLength(algorithm string) int {
	switch algorithm {
	case "md5":
		return 32
	case "sha256":
		return 64
	default:
		return 0
	}
}
