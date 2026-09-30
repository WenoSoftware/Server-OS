package state

import (
	"sync"
)

// ServerState tracks our background node status securely
type ServerState struct {
	Mu             sync.RWMutex `json:"-"`
	SharingEnabled bool         `json:"sharing_enabled"`
	Version        string       `json:"version"`
}

// Global state instance for the daemon
var State = ServerState{
	SharingEnabled: false,
	Version:        "0.1.0",
}