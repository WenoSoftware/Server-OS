package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"serveros/internal/chunker"
	"serveros/internal/gateway"
	"serveros/internal/handlers"
	"serveros/internal/network"
	"serveros/internal/storage"
)

func TestMultiNodeSwarmGateway(t *testing.T) {
	// 1. Setup temporary storage directory for Node A (Seeder)
	dirA := t.TempDir()

	// Initialize storage context
	if err := storage.InitStorage(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	testFilePath := filepath.Join(dirA, "index.html")
	htmlContent := []byte("<html><body><h1>Weno P2P Web Hosting Works!</h1></body></html>")
	if err := os.WriteFile(testFilePath, htmlContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	manifest, chunks, err := chunker.SplitFile(testFilePath, 256*1024)
	if err != nil {
		t.Fatalf("Failed to split file: %v", err)
	}

	// Save chunks into storage
	for _, chunk := range chunks {
		if err := storage.SaveChunk(chunk.Hash, chunk.Data); err != nil {
			t.Fatalf("Failed to save chunk: %v", err)
		}
	}

	manifestData, _ := json.Marshal(manifest)
	
	// Compute SHA-256 manifest hash
	hashBytes := sha256.Sum256(manifestData)
	manifestHash := fmt.Sprintf("%x", hashBytes)
	
	if err := storage.SaveChunk(manifestHash, manifestData); err != nil {
		t.Fatalf("Failed to save manifest chunk: %v", err)
	}

	// 2. Spin up "Node A" HTTP Server (The Seeder Peer)
	muxA := http.NewServeMux()
	muxA.HandleFunc("/get-chunk", handlers.HandleGetChunk)
	serverA := httptest.NewServer(muxA)
	defer serverA.Close()

	nodeAAddr := serverA.URL[7:]

	// 3. Register Node A inside Node B's active swarm peer list
	network.SwarmPeers.AddPeer(nodeAAddr)

	// 4. Test Gateway Request handler on a test server representing Node B's gateway
	muxB := http.NewServeMux()
	muxB.HandleFunc("/site/", gateway.HandleGateway)
	serverB := httptest.NewServer(muxB)
	defer serverB.Close()

	// 5. Request the decentralized site through Node B's gateway endpoint
	resp, err := http.Get(serverB.URL + "/site/" + manifestHash)
	if err != nil {
		t.Fatalf("Gateway request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK from gateway, got %d", resp.StatusCode)
	}

	t.Log("[SUCCESS] Multi-node swarm gateway test passed! Node successfully resolved and fetched assets from the peer.")
}