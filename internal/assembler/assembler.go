package assembler

import (
	"fmt"
	"os"
	"path/filepath"

	"serveros/internal/chunker"
	"serveros/internal/storage"
)

// ReassembleFile reads a file manifest, retrieves all necessary chunks from local storage,
// and reconstructs the original file into the target output directory.
func ReassembleFile(manifest *chunker.FileManifest, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	outputPath := filepath.Join(outputDir, manifest.Filename)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create target file: %v", err)
	}
	defer outFile.Close()

	// Iterate through every chunk hash declared in the manifest blueprint
	for i, hash := range manifest.ChunkHashes {
		// Load chunk bytes from local CAS storage directory
		chunkData, err := storage.LoadChunk(hash)
		if err != nil {
			return "", fmt.Errorf("missing required chunk %d (%s) from storage: %v", i, hash[:8], err)
		}

		// Write chunk data sequentially into the output file stream
		if _, err := outFile.Write(chunkData); err != nil {
			return "", fmt.Errorf("failed to write chunk %d to file: %v", i, err)
		}
	}

	fmt.Printf("[ASSEMBLER] Successfully reassembled '%s' from %d verified chunks!\n", manifest.Filename, len(manifest.ChunkHashes))
	return outputPath, nil
}