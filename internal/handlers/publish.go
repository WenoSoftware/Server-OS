package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"serveros/internal/chunker"
	"serveros/internal/storage"
)

// HandlePublish accepts a file via HTTP POST, chunks it, and saves it to local CAS storage
func HandlePublish(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for local web applications connecting via the SDK
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32MB upload)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing 'file' field in form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary file on disk to pass through your chunker
	tempFile, err := os.CreateTemp("", "serveros-upload-*")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Failed to write temp file", http.StatusInternalServerError)
		return
	}

	// Use your existing chunker engine
	manifest, chunks, err := chunker.SplitFile(tempFile.Name(), 256*1024)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to split file: %v", err), http.StatusInternalServerError)
		return
	}

	// Override manifest filename with the original upload name
	manifest.Filename = header.Filename

	// Save individual chunks to CAS
	for _, chunk := range chunks {
		if err := storage.SaveChunk(chunk.Hash, chunk.Data); err != nil {
			http.Error(w, "Failed to save chunk to storage", http.StatusInternalServerError)
			return
		}
	}

	// Marshal and save the manifest into CAS
	manifestData, _ := json.Marshal(manifest)
	hashBytes := sha256.Sum256(manifestData)
	manifestHash := fmt.Sprintf("%x", hashBytes)

	if err := storage.SaveChunk(manifestHash, manifestData); err != nil {
		http.Error(w, "Failed to save manifest chunk", http.StatusInternalServerError)
		return
	}

	// Respond with the resulting manifest hash JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"manifestHash": manifestHash,
		"filename":     header.Filename,
	})
}