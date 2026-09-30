package network

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// PeerManager tracks active nodes participating in the swarm
type PeerManager struct {
	mu    sync.RWMutex
	Peers map[string]bool // Mapping of peer address (e.g., "127.0.0.1:8081") to active status
}

// Global swarm peer registry
var SwarmPeers = PeerManager{
	Peers: make(map[string]bool),
}

// AddPeer registers a new verified peer node into the network swarm
func (pm *PeerManager) AddPeer(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.Peers[address] = true
	fmt.Printf("[NETWORK] Added peer to swarm: %s\n", address)
}

// RemovePeer drops a node from the swarm
func (pm *PeerManager) RemovePeer(address string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.Peers, address)
	fmt.Printf("[NETWORK] Removed peer from swarm: %s\n", address)
}

// FetchChunkFromPeer connects to a peer daemon, downloads a chunk, and instantly verifies its hash
// FetchChunkFromPeer connects to a peer daemon, downloads a chunk, and instantly verifies its hash
func FetchChunkFromPeer(peerAddress string, expectedHash string) ([]byte, error) {
	// Use the hash parameter to query the CAS storage node
	url := fmt.Sprintf("http://%s/get-chunk?hash=%s", peerAddress, expectedHash)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %v", peerAddress, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer responded with status code: %d", resp.StatusCode)
	}

	chunkData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read incoming chunk stream: %v", err)
	}

	// Cryptographic verification firewall check
	hasher := sha256.New()
	hasher.Write(chunkData)
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))

	if calculatedHash != expectedHash {
		return nil, errors.New("security warning: peer sent a tampered chunk (hash mismatch rejected)")
	}

	return chunkData, nil
}