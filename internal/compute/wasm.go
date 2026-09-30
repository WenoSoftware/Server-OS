package compute

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// RunWasmFunction executes a compiled Wasm binary loaded from CAS
func RunWasmFunction(wasmBytes []byte, w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	rnt := wazero.NewRuntime(ctx)
	defer rnt.Close(ctx)

	// Instantiate WASI support so the Wasm module can read/write I/O
	wasi_snapshot_preview1.MustInstantiate(ctx, rnt)

	// Configure stdout/stderr capture to return the function's output to the HTTP client
	compiled, err := rnt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile wasm module: %v", err)
	}

	config := wazero.NewModuleConfig().WithStdout(w).WithStderr(w)
	_, err = rnt.InstantiateModule(ctx, compiled, config)
	if err != nil {
		return fmt.Errorf("failed to execute wasm module: %v", err)
	}

	return nil
}