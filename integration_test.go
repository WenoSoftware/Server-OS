package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"serveros/internal/chunker"
	"serveros/internal/handlers"
	"serveros/internal/network"
	"serveros/internal/state"
	"serveros/internal/storage"
)

func TestWenoEndToEndIntegration(t *testing.T) {
	// 1. Initialize storage and enable master switch for testing
	if err := storage.InitStorage(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	state.State.SharingEnabled = true

	// 2. Create a dummy test file in a temp directory
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "sample.txt")
	testData := []byte("Weno Server OS decentralized storage engine integration test payload content!")
	if err := os.WriteFile(testFilePath, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// 3. Test Chunker: Split file into chunks (using 256 KB block size)
	manifest, chunks, err := chunker.SplitFile(testFilePath, 256*1024)
	if err != nil {
		t.Fatalf("Failed to split file: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected chunks to be generated, got 0")
	}

	t.Logf("[CHUNKER] Successfully split %s into %d chunks", manifest.Filename, len(chunks))

	// 4. Save chunks to local storage engine
	targetHash := chunks[0].Hash
	if err := storage.SaveChunk(targetHash, chunks[0].Data); err != nil {
		t.Fatalf("Failed to save chunk to storage: %v", err)
	}

	// 5. Setup a test HTTP server utilizing your handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/get-chunk", handlers.HandleGetChunk)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 6. Test Network Client: Fetch and verify chunk from the node
	peerAddr := ts.URL[7:] // Strip "http://" prefix to get host:port

	fetchedData, err := network.FetchChunkFromPeer(peerAddr, targetHash)
	if err != nil {
		t.Fatalf("Failed to fetch chunk via network client: %v", err)
	}

	if string(fetchedData) != string(chunks[0].Data) {
		t.Errorf("Data mismatch! Expected %s, got %s", string(chunks[0].Data), string(fetchedData))
	}

	t.Log("[SUCCESS] End-to-end integration test passed!")
}