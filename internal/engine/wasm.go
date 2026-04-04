package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runner encapsula a lógica de execução do módulo WebAssembly.
type Runner struct {
	runtime wazero.Runtime
}

// NewRunner inicializa o runtime do wazero.
func NewRunner(ctx context.Context) (*Runner, error) {
	r := wazero.NewRuntime(ctx)

	// Instanciar WASI para permitir que o módulo interaja com o sistema de arquivos, se necessário.
	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &Runner{runtime: r}, nil
}

// RunFunction carrega um arquivo Wasm, instancia o módulo e executa uma função exportada.
func (r *Runner) RunFunction(ctx context.Context, wasmPath string, funcName string, params ...uint64) ([]uint64, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}

	// Compila o módulo Wasm. O wazero valida o arquivo aqui.
	mod, err := r.runtime.Instantiate(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate wasm module (is it a valid Wasm file?): %w", err)
	}
	defer mod.Close(ctx)

	// Obtém a função exportada.
	f := mod.ExportedFunction(funcName)
	if f == nil {
		return nil, fmt.Errorf("function %s not found in module", funcName)
	}

	// Executa a função.
	results, err := f.Call(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to call function %s: %w", funcName, err)
	}

	return results, nil
}

// Close libera os recursos do runtime.
func (r *Runner) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}
