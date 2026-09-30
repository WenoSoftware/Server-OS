package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// StorageDir defines the local directory where Weno nodes cache network chunks
const StorageDir = "./weno_storage"

// InitStorage ensures the local storage directory exists on disk
func InitStorage() error {
	if err := os.MkdirAll(StorageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}
	return nil
}

// SaveChunk writes a verified chunk to disk using its SHA-256 hash as the filename
func SaveChunk(hash string, data []byte) error {
	if err := InitStorage(); err != nil {
		return err
	}

	filePath := filepath.Join(StorageDir, hash)
	
	// Write the raw chunk data to disk with secure permissions
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save chunk %s to disk: %v", hash[:8], err)
	}

	fmt.Printf("[STORAGE] Saved chunk to disk (CAS): %s\n", hash[:8])
	return nil
}

// GetChunk reads a chunk's binary data from local CAS storage by hash
func GetChunk(hash string) ([]byte, error) {
	filePath := filepath.Join(StorageDir, hash)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("chunk not found in local storage: %s", hash[:8])
	}

	return data, nil
}

// LoadChunk is an alias for GetChunk for backward compatibility with the assembler
func LoadChunk(hash string) ([]byte, error) {
	return GetChunk(hash)
}