package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"serveros/internal/network"
)

func TestSwarmChunkExchange(t *testing.T) {
	// 1. Create a test HTTP server representing "Peer B" in the swarm
	mux := http.NewServeMux()
	
	expectedContent := "hello-weno-decentralized-world"
	
	// Compute the correct hash for this content
	hasher := sha256.New()
	hasher.Write([]byte(expectedContent))
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	// Register the chunk serving handler on Peer B
	mux.HandleFunc("/get-chunk", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	})

	// Start a local test server representing the peer node
	peerServer := httptest.NewServer(mux)
	defer peerServer.Close()

	// Extract just the host:port from the test server URL (e.g., "127.0.0.1:12345")
	peerAddr := peerServer.Listener.Addr().String()

	t.Logf("[TEST] Peer B active at %s", peerAddr)

	// 2. Simulate "Node A" fetching the chunk from Peer B using our network client
	chunkData, err := network.FetchChunkFromPeer(peerAddr, correctHash)
	if err != nil {
		t.Fatalf("[FAIL] Failed to fetch and verify chunk from peer: %v", err)
	}

	// 3. Validate the results
	receivedContent := string(chunkData)
	if receivedContent != expectedContent {
		t.Errorf("[FAIL] Expected content '%s', got '%s'", expectedContent, receivedContent)
	}

	t.Logf("[SUCCESS] Node A successfully fetched and verified chunk: '%s'", receivedContent)
}

func TestSwarmTamperDetection(t *testing.T) {
	// Test that our security firewall catches malicious or corrupted chunks from a bad peer
	mux := http.NewServeMux()
	
	maliciousContent := "tampered-malicious-payload"
	mux.HandleFunc("/get-chunk", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(maliciousContent))
	})

	peerServer := httptest.NewServer(mux)
	defer peerServer.Close()
	peerAddr := peerServer.Listener.Addr().String()

	// Give it a fake/wrong expected hash to trigger the security firewall
	fakeExpectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	_, err := network.FetchChunkFromPeer(peerAddr, fakeExpectedHash)
	if err == nil {
		t.Fatal("[FAIL] Security firewall failed to catch tampered chunk!")
	}

	t.Logf("[SUCCESS] Security firewall successfully blocked tampered chunk: %v", err)
}