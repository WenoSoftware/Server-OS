package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"serveros/internal/chunker"
	"serveros/internal/discovery"
	"serveros/internal/gateway"
	"serveros/internal/handlers"
	"serveros/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		port := "8080"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		startDaemon(port)

	case "share":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing file path. Usage: go run . share <filepath>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		shareFile(filePath)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Weno Server OS - Decentralized P2P Cloud Infrastructure")
	fmt.Println("Usage:")
	fmt.Println("  go run . start [port]        - Start the background daemon & auto-discovery")
	fmt.Println("  go run . share <filepath>    - Split a file and seed its chunks into CAS storage")
}

func startDaemon(port string) {
	fmt.Printf("Starting Weno Server OS Daemon on port %s...\n", port)

	if err := storage.InitStorage(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Register HTTP routes
	http.HandleFunc("/status", handlers.HandleStatus)
	http.HandleFunc("/toggle", handlers.HandleToggle)
	http.HandleFunc("/verify-chunk", handlers.HandleVerifyChunk)
	http.HandleFunc("/get-chunk", handlers.HandleGetChunk)
	http.HandleFunc("/publish", handlers.HandlePublish) // 👈 Add this line
	http.HandleFunc("/site/", gateway.HandleGateway)

	// Start background mDNS-style UDP peer discovery
	go discovery.StartBroadcaster(port)
	go discovery.StartListener(port)

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	fmt.Printf("Daemon listening securely on http://%s\n", addr)
	fmt.Println("[INFO] Background peer discovery active.")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Daemon failed to start: %v", err)
	}
}

func shareFile(filePath string) {
	if err := storage.InitStorage(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	manifest, chunks, err := chunker.SplitFile(filePath, 256*1024)
	if err != nil {
		log.Fatalf("Failed to split file: %v", err)
	}

	fmt.Printf("[CHUNKER] Sliced '%s' into %d cryptographic chunks.\n", manifest.Filename, len(chunks))

	for _, chunk := range chunks {
		if err := storage.SaveChunk(chunk.Hash, chunk.Data); err != nil {
			log.Printf("Failed to save chunk %s: %v", chunk.Hash[:8], err)
		}
	}

	// Marshal and save the manifest itself into CAS
	manifestData, _ := json.Marshal(manifest)
	hashBytes := sha256.Sum256(manifestData)
	manifestHash := fmt.Sprintf("%x", hashBytes)
	
	if err := storage.SaveChunk(manifestHash, manifestData); err != nil {
		log.Fatalf("Failed to save manifest chunk: %v", err)
	}

	fmt.Println("[SUCCESS] All chunks successfully committed to local CAS.")
	fmt.Printf("🌐 Manifest Hash (URL Key): %s\n", manifestHash)
}