package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"serveros/internal/state"
	"serveros/internal/storage"
)

// VerifyRequest represents an incoming chunk and its expected hash
type VerifyRequest struct {
	ExpectedHash string `json:"expected_hash"`
	Content      string `json:"content"`
}

// HandleStatus reports daemon health and sharing state
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	defer state.State.Mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.State)
}

// HandleToggle controls the master on/off switch
func HandleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state.State.Mu.Lock()
	state.State.SharingEnabled = !state.State.SharingEnabled
	currentState := state.State.SharingEnabled
	state.State.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"sharing_enabled": currentState,
	})
	
	if currentState {
		fmt.Println("[STATUS] Master Switch: ON - Background sharing active.")
	} else {
		fmt.Println("[STATUS] Master Switch: OFF - Background sharing halted.")
	}
}

// HandleVerifyChunk processes the cryptographic security firewall check
func HandleVerifyChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state.State.Mu.RLock()
	if !state.State.SharingEnabled {
		state.State.Mu.RUnlock()
		http.Error(w, "Sharing is disabled via master switch", http.StatusForbidden)
		return
	}
	state.State.Mu.RUnlock()

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Compute SHA-256 hash of the incoming content chunk
	hasher := sha256.New()
	hasher.Write([]byte(req.Content))
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))

	w.Header().Set("Content-Type", "application/json")

	if calculatedHash == req.ExpectedHash {
		fmt.Printf("[SECURITY] VERIFIED: Chunk hash match confirmed (%s)\n", calculatedHash[:8])
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"verified": true,
			"message":  "Cryptographic check passed. Chunk accepted.",
		})
	} else {
		fmt.Printf("[SECURITY] REJECTED: Hash mismatch! Expected %s, got %s\n", req.ExpectedHash[:8], calculatedHash[:8])
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"verified": false,
			"message":  "Security warning: Hash mismatch. Chunk rejected.",
		})
	}
}

// HandleGetChunk serves a local file chunk from disk to a requesting peer in the swarm
func HandleGetChunk(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	if !state.State.SharingEnabled {
		state.State.Mu.RUnlock()
		http.Error(w, "Sharing is disabled via master switch", http.StatusForbidden)
		return
	}
	state.State.Mu.RUnlock()

	chunkHash := r.URL.Query().Get("hash")
	if chunkHash == "" {
		http.Error(w, "Missing chunk hash parameter", http.StatusBadRequest)
		return
	}

	// Load the chunk data from our local CAS storage directory via the storage package
	chunkData, err := storage.LoadChunk(chunkHash)
	if err != nil {
		http.Error(w, "Chunk not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(chunkData)

	fmt.Printf("[NETWORK] Served chunk %s to a requesting peer.\n", chunkHash[:8])
}