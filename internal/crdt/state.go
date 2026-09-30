package crdt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Record struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	AuthorID  string `json:"author_id"`
	Signature []byte `json:"signature"` // Cryptographic signature of the record
}

type DistributedState struct {
	mu      sync.RWMutex
	Storage map[string]Record `json:"storage"`
	privKey crypto.PrivKey
}

func NewState() *DistributedState {
	return &DistributedState{
		Storage: make(map[string]Record),
	}
}

// SetPrivKey assigns the node's private key for signing local writes
func (ds *DistributedState) SetPrivKey(privKey crypto.PrivKey) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.privKey = privKey
}

// Set updates a key locally with a timestamp and cryptographic signature
func (ds *DistributedState) Set(key, value, authorID string) (Record, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	timestamp := time.Now().UnixNano()
	var sig []byte
	var err error

	if ds.privKey != nil {
		payload := fmt.Sprintf("%s:%s:%d:%s", key, value, timestamp, authorID)
		sig, err = ds.privKey.Sign([]byte(payload))
		if err != nil {
			return Record{}, fmt.Errorf("failed to sign record: %v", err)
		}
	}

	rec := Record{
		Value:     value,
		Timestamp: timestamp,
		AuthorID:  authorID,
		Signature: sig,
	}
	ds.Storage[key] = rec
	return rec, nil
}

// verifyRecord checks if the cryptographic signature matches the author ID
func verifyRecord(key string, rec Record) bool {
	if len(rec.Signature) == 0 {
		return false // Reject unsigned records for strict security
	}

	pid, err := peer.Decode(rec.AuthorID)
	if err != nil {
		return false
	}
	
	pubKey, err := pid.ExtractPublicKey()
	if pubKey == nil || err != nil {
		return false
	}

	payload := fmt.Sprintf("%s:%s:%d:%s", key, rec.Value, rec.Timestamp, rec.AuthorID)
	valid, err := pubKey.Verify([]byte(payload), rec.Signature)
	return err == nil && valid
}

// Merge takes incoming peer state, validates signatures, and merges via Last-Write-Wins (LWW)
func (ds *DistributedState) Merge(incomingJSON []byte) error {
	var incoming map[string]Record
	if err := json.Unmarshal(incomingJSON, &incoming); err != nil {
		return err
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()

	for k, incRec := range incoming {
		// Drop untrusted or forged records immediately
		if !verifyRecord(k, incRec) {
			continue
		}

		locRec, exists := ds.Storage[k]
		if !exists || incRec.Timestamp > locRec.Timestamp {
			ds.Storage[k] = incRec
		}
	}
	return nil
}

// Export serializes the entire state map to broadcast across the swarm
func (ds *DistributedState) Export() ([]byte, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return json.Marshal(ds.Storage)
}

// Get retrieves a record from the distributed state safely in a thread-safe manner
func (ds *DistributedState) Get(key string) (Record, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	rec, exists := ds.Storage[key]
	return rec, exists
}