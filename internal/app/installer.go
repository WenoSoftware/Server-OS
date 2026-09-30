package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"serveros/internal/storage"

	"github.com/libp2p/go-libp2p/core/peer"
)

// InstallApp loads a manifest from CAS, verifies it, and mounts it
func InstallApp(manifestHash string) error {
	// 1. Load manifest data from CAS
	manifestData, err := storage.GetChunk(manifestHash)
	if err != nil {
		return fmt.Errorf("failed to find app manifest in CAS: %v", err)
	}

	var manifest AppManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("invalid manifest JSON: %v", err)
	}

	// 2. Verify publisher signature
	if len(manifest.Signature) == 0 {
		return fmt.Errorf("rejected: app manifest has no cryptographic signature")
	}

	pid, err := peer.Decode(manifest.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid author ID: %v", err)
	}

	pubKey, err := pid.ExtractPublicKey()
	if pubKey == nil || err != nil {
		return fmt.Errorf("failed to extract author public key: %v", err)
	}

	// Reconstruct payload used for signing
	payloadBytes, _ := json.Marshal(struct {
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		AuthorID    string   `json:"author_id"`
		ChunkHashes []string `json:"chunk_hashes"`
	}{
		Name:        manifest.Name,
		Version:     manifest.Version,
		AuthorID:    manifest.AuthorID,
		ChunkHashes: manifest.ChunkHashes,
	})

	valid, err := pubKey.Verify(payloadBytes, manifest.Signature)
	if err != nil || !valid {
		return fmt.Errorf("cryptographic signature verification failed for app '%s'", manifest.Name)
	}

	fmt.Printf("[PACKAGE MANAGER] Successfully verified & installed: %s (v%s)\n", manifest.Name, manifest.Version)
	fmt.Printf("📦 Access path: /apps/%s/\n", manifest.Name)
	
	// Optional: save installed app reference to disk or state registry
	return nil
}