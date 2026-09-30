package app

type AppManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	AuthorID    string   `json:"author_id"`
	Entrypoint  string   `json:"entrypoint"` // e.g., "index.html"
	ChunkHashes []string `json:"chunk_hashes"`
	Signature   []byte   `json:"signature"`
}