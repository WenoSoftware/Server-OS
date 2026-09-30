package cloud

import (
	"io"
	"net/http"
	"strings"

	"serveros/internal/chunker"
	"serveros/internal/storage"
)

// HandleS3Storage mimics a basic S3 API route
func HandleS3Storage(w http.ResponseWriter, r *http.Request) {
	// Format: /s3/<bucket>/<key>
	path := strings.TrimPrefix(r.URL.Path, "/s3/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "Invalid S3 path format: /s3/{bucket}/{key}", http.StatusBadRequest)
		return
	}

	bucket, key := parts[0], parts[1]
	_ = bucket // Buckets can be mapped to namespaces if desired

	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		// Save directly into CAS using chunking or raw blob storage
		hash, err := storage.SaveBlobDirect(key, body)
		if err != nil {
			http.Error(w, "Failed to save object", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","etag":"` + hash + `"}`))

	case http.MethodGet:
		data, err := storage.GetChunk(key) // or lookup via key mapping
		if err != nil {
			http.Error(w, "Object not found", http.StatusNotFound)
			return
		}
		w.Write(data)

	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}