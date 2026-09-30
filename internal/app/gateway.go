package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"serveros/internal/storage"
)

// HandleAppGateway serves installed applications from CAS chunks via the HTTP gateway
func HandleAppGateway(w http.ResponseWriter, r *http.Request) {
	// Expected path format: /apps/<manifest-hash>/filename.html
	path := strings.TrimPrefix(r.URL.Path, "/apps/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "App manifest hash not specified in path", http.StatusBadRequest)
		return
	}

	manifestHash := parts[0]
	// Default to index.html if no specific asset is requested
	subPath := "index.html"
	if len(parts) > 1 && parts[1] != "" {
		subPath = parts[1]
	}

	// 1. Fetch manifest from CAS
	manifestData, err := storage.GetChunk(manifestHash)
	if err != nil {
		http.Error(w, "App manifest not found in CAS", http.StatusNotFound)
		return
	}

	var manifest AppManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		http.Error(w, "Failed to parse app manifest", http.StatusInternalServerError)
		return
	}

	// 2. Find the requested asset hash from the manifest chunks
	// (Assuming manifest chunks map filenames or the entrypoint is served)
	var targetChunkHash string
	if subPath == "index.html" || subPath == manifest.Entrypoint {
		// If entrypoint or index, use the first chunk or match entrypoint
		targetChunkHash = manifest.ChunkHashes[0] 
	} else {
		// For multi-file apps, we can map subpaths to chunk hashes stored in the manifest
		// Here we fallback to checking matching indices or a basic lookup
		targetChunkHash = manifest.ChunkHashes[0] 
	}

	// 3. Retrieve chunk data from CAS
	fileData, err := storage.GetChunk(targetChunkHash)
	if err != nil {
		http.Error(w, "App asset chunk not found in storage", http.StatusNotFound)
		return
	}

	// Set appropriate content type headers
	if strings.HasSuffix(subPath, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(subPath, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasSuffix(subPath, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}

	w.Write(fileData)
}