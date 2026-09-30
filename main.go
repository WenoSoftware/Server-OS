package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"serveros/internal/app"
	"serveros/internal/chunker"
	"serveros/internal/cloud"
	"serveros/internal/compute"
	"serveros/internal/crdt"
	"serveros/internal/discovery"
	"serveros/internal/gateway"
	"serveros/internal/handlers"
	"serveros/internal/p2p"
	"serveros/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 1. Initialize the P2P networking node globally
	p2pHost, err := p2p.InitP2PNode()
	if err != nil {
		log.Fatalf("Failed to start P2P node: %v", err)
	}
	defer p2pHost.Close()

	// 2. Set up the wide-area libp2p chunk stream protocol handler
	p2p.SetupChunkProtocol(p2pHost)

	// 3. Initialize Pub/Sub manager for real-time node messaging
	ctx := context.Background()
	psManager, err := p2p.InitPubSub(ctx, p2pHost)
	if err != nil {
		log.Fatalf("Failed to initialize pubsub: %v", err)
	}

	// 4. Start the CRDT state synchronization engine over Pub/Sub
	crdtState := crdt.StartCRDTSync(ctx, psManager)

	// 5. Provide the node's cryptographic private key to the state engine for signing writes
	if privKey := p2pHost.Peerstore().PrivKey(p2pHost.ID()); privKey != nil {
		crdtState.SetPrivKey(privKey)
	}

	command := os.Args[1]

	switch command {
	case "start":
		port := "8080"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		startDaemon(port, crdtState, p2pHost.ID().String())

	case "share":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing file path. Usage: go run . share <filepath>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		shareFile(filePath)

	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing manifest hash. Usage: go run . install <manifest-hash>")
			os.Exit(1)
		}
		manifestHash := os.Args[2]
		if err := storage.InitStorage(); err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
		if err := app.InstallApp(manifestHash); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Weno Server OS - Sovereign Decentralized Cloud")
	fmt.Println("Usage:")
	fmt.Println("  go run . start [port]        - Start the background daemon, DHT & cloud services")
	fmt.Println("  go run . share <filepath>    - Split a file and seed its chunks into CAS storage")
	fmt.Println("  go run . install <hash>      - Install and verify a decentralized app bundle")
}

func startDaemon(port string, crdtState *crdt.DistributedState, authorID string) {
	fmt.Printf("Starting Weno Server OS Cloud Daemon on port %s...\n", port)

	if err := storage.InitStorage(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Core System & Gateway Routes
	http.HandleFunc("/status", handlers.HandleStatus)
	http.HandleFunc("/toggle", handlers.HandleToggle)
	http.HandleFunc("/verify-chunk", handlers.HandleVerifyChunk)
	http.HandleFunc("/get-chunk", handlers.HandleGetChunk)
	http.HandleFunc("/publish", handlers.HandlePublish)
	http.HandleFunc("/site/", gateway.HandleGateway)

	// Decentralized Package Manager App Routes
	http.HandleFunc("/apps/", app.HandleAppGateway)

	// Decentralized Cloud Primitives
	// 1. S3-Compatible Object Storage API
	http.HandleFunc("/s3/", cloud.HandleS3Storage)

	// 2. Serverless Wasm Compute Runtime
	http.HandleFunc("/compute/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/compute/")
		wasmBytes, err := storage.GetChunk(path)
		if err != nil {
			http.Error(w, "Wasm binary not found in CAS storage", http.StatusNotFound)
			return
		}
		if err := compute.RunWasmFunction(wasmBytes, w, r); err != nil {
			http.Error(w, fmt.Sprintf("Wasm execution error: %v", err), http.StatusInternalServerError)
		}
	})

	// Decentralized Database (CRDT) HTTP routes
	dbHandler := &handlers.DBHandler{
		State:    crdtState,
		AuthorID: authorID,
	}
	http.HandleFunc("/db/get", dbHandler.HandleGet)
	http.HandleFunc("/db/set", dbHandler.HandleSet)

	// Start background local mDNS discovery & global Kademlia DHT routing
	go discovery.StartBroadcaster(port)
	go discovery.StartListener(port)

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	fmt.Printf("☁️ Cloud Daemon listening securely on http://%s\n", addr)
	fmt.Println("[INFO] Active services: CAS, S3 Storage, Wasm Serverless, CRDT Database, mDNS, & Global DHT.")

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