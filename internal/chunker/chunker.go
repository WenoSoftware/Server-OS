package chunker

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// Chunk represents a single piece of a file with its cryptographic proof
type Chunk struct {
	Index int    `json:"index"`
	Hash  string `json:"hash"`
	Data  []byte `json:"data"`
}

// FileManifest represents the blueprint for a shared file or website asset
type FileManifest struct {
	Filename   string   `json:"filename"`
	TotalSize  int64    `json:"total_size"`
	ChunkSize  int      `json:"chunk_size"`
	ChunkHashes []string `json:"chunk_hashes"`
}

// SplitFile reads a file, breaks it into fixed-size chunks, and generates a manifest
func SplitFile(filePath string, chunkSize int) (*FileManifest, []Chunk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	manifest := &FileManifest{
		Filename:    fileInfo.Name(),
		TotalSize:   fileInfo.Size(),
		ChunkSize:   chunkSize,
		ChunkHashes: []string{},
	}

	var chunks []Chunk
	buffer := make([]byte, chunkSize)
	index := 0

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}
		if n == 0 {
			break
		}

		// Copy the read bytes into a dedicated chunk buffer
		chunkData := make([]byte, n)
		copy(chunkData, buffer[:n])

		// Calculate SHA-256 hash for this specific chunk
		hasher := sha256.New()
		hasher.Write(chunkData)
		chunkHash := hex.EncodeToString(hasher.Sum(nil))

		manifest.ChunkHashes = append(manifest.ChunkHashes, chunkHash)
		chunks = append(chunks, Chunk{
			Index: index,
			Hash:  chunkHash,
			Data:  chunkData,
		})

		index++
	}

	return manifest, chunks, nil
}