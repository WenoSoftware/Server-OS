package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"serveros/internal/assembler"
	"serveros/internal/chunker"
	"serveros/internal/network"
	"serveros/internal/storage"
)

// HandleGateway processes incoming browser requests for decentralized sites/files.
// Expected URL pattern: /site/<manifest-hash>/<optional-file-path>
func HandleGateway(w http.ResponseWriter, r *http.Request) {
	// Enable CORS so local web applications can talk to the daemon freely
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/site/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		http.Error(w, "Bad Request: Missing manifest hash in URL path", http.StatusBadRequest)
		return
	}

	manifestHash := pathParts[0]

	// 1. Load the manifest chunk from local storage (or fetch it from peers if missing)
	manifestData, err := storage.LoadChunk(manifestHash)
	if err != nil {
		// Try fetching manifest from active swarm peers if not found locally
		fetched, fetchErr := fetchManifestFromSwarm(manifestHash)
		if fetchErr != nil {
			http.Error(w, fmt.Sprintf("Manifest not found locally or in swarm: %v", manifestHash), http.StatusNotFound)
			return
		}
		manifestData = fetched
	}

	var manifest chunker.FileManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		http.Error(w, "Failed to parse site manifest blueprint", http.StatusInternalServerError)
		return
	}

	// 2. Ensure all chunks referenced in the manifest are available locally
	for _, chunkHash := range manifest.ChunkHashes {
		// Check if chunk exists locally by attempting to load it
		if _, err := storage.LoadChunk(chunkHash); err != nil {
			// Missing chunk! Attempt to pull it from the swarm network
			err := pullChunkFromSwarm(chunkHash)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to retrieve required chunk %s from swarm", chunkHash[:8]), http.StatusServiceUnavailable)
				return
			}
		}
	}

	// 3. Reassemble the site temporarily in memory/cache to serve it
	tempDir := filepath.Join(os.TempDir(), "weno_cache", manifestHash)
	outputPath, err := assembler.ReassembleFile(&manifest, tempDir)
	if err != nil {
		http.Error(w, "Failed to reassemble decentralized asset", http.StatusInternalServerError)
		return
	}

	// 4. Serve the file/page back to the browser
	http.ServeFile(w, r, outputPath)
}

// Helper to find and pull a manifest chunk from connected peers
func fetchManifestFromSwarm(manifestHash string) ([]byte, error) {
	// Loop over the map keys (peer addresses are strings)
	for peerAddr := range network.SwarmPeers.Peers {
		data, err := network.FetchChunkFromPeer(peerAddr, manifestHash)
		if err == nil {
			// Cache it locally once found
			_ = storage.SaveChunk(manifestHash, data)
			return data, nil
		}
	}
	return nil, fmt.Errorf("manifest unavailable across all connected peers")
}

// Helper to pull a missing content chunk from the swarm
func pullChunkFromSwarm(chunkHash string) error {
	for peerAddr := range network.SwarmPeers.Peers {
		data, err := network.FetchChunkFromPeer(peerAddr, chunkHash)
		if err == nil {
			return storage.SaveChunk(chunkHash, data)
		}
	}
	return fmt.Errorf("chunk %s not found on any peer", chunkHash[:8])
}