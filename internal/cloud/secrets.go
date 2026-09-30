package cloud

import (
	"encoding/base64"
	"fmt"
	
	"serveros/internal/crdt"
)

// SetSecret stores an encrypted config parameter into the CRDT database
func SetSecret(state *crdt.DistributedState, secretName, secretValue string) error {
	// In a full implementation, you'd encrypt secretValue using the target peer's public key
	encodedVal := base64.StdEncoding.EncodeToString([]byte(secretValue))
	
	key := fmt.Sprintf("secret/%s", secretName)
	_, err := state.Set(key, encodedVal, "system-admin")
	return err
}

// GetSecret retrieves and decodes a secret parameter
func GetSecret(state *crdt.DistributedState, secretName string) (string, error) {
	key := fmt.Sprintf("secret/%s", secretName)
	rec, exists := state.Get(key)
	if !exists {
		return "", fmt.Errorf("secret not found")
	}

	decoded, err := base64.StdEncoding.DecodeString(rec.Value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}