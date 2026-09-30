package handlers

import (
	"encoding/json"
	"net/http"

	"serveros/internal/crdt"
)

type DBHandler struct {
	State    *crdt.DistributedState
	AuthorID string
}

// HandleGet handles GET /db/get?key=foo
func (h *DBHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing 'key' query parameter", http.StatusBadRequest)
		return
	}

	rec, exists := h.State.Get(key)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "key not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rec)
}

// HandleSet handles POST /db/set with JSON body {"key": "...", "value": "..."}
func (h *DBHandler) HandleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update local CRDT state (which will automatically sync to peers via Pub/Sub)
	rec := h.State.Set(req.Key, req.Value, h.AuthorID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"record":  rec,
	})
}