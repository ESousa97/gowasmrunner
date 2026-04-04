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
	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	return &Runner{runtime: r}, nil
}

// RunFunction carrega um arquivo Wasm e executa uma função numérica genérica.
func (r *Runner) RunFunction(ctx context.Context, wasmPath string, funcName string, params ...uint64) ([]uint64, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}

	mod, err := r.runtime.Instantiate(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate wasm module: %w", err)
	}
	defer mod.Close(ctx)

	f := mod.ExportedFunction(funcName)
	if f == nil {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	return f.Call(ctx, params...)
}

// RunGreet carrega o módulo e executa a lógica de saudação com strings.
func (r *Runner) RunGreet(ctx context.Context, wasmPath string, name string) (string, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return "", fmt.Errorf("failed to read wasm file: %w", err)
	}

	mod, err := r.runtime.Instantiate(ctx, wasmBytes)
	if err != nil {
		return "", fmt.Errorf("failed to instantiate wasm module: %w", err)
	}
	defer mod.Close(ctx)

	// 1. Obter funções exportadas
	greetFunc := mod.ExportedFunction("greet")
	allocFunc := mod.ExportedFunction("allocate")
	if greetFunc == nil || allocFunc == nil {
		return "", fmt.Errorf("exported functions 'greet' or 'allocate' not found")
	}

	// 2. Alocar memória no Guest para a string de entrada
	nameSize := uint64(len(name))
	results, err := allocFunc.Call(ctx, nameSize)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}
	namePtr := results[0]

	// 3. Escrever a string na memória linear do Guest
	// O wazero nos dá acesso direto ao slice de bytes da memória do Wasm.
	if !mod.Memory().Write(uint32(namePtr), []byte(name)) {
		return "", fmt.Errorf("out of memory bounds when writing string")
	}

	// 4. Chamar a função greet(ptr, len)
	// O Guest retornará um uint64 que contém o ponteiro (32 bits superiores) e o tamanho (32 bits inferiores) do resultado.
	greetResults, err := greetFunc.Call(ctx, namePtr, nameSize)
	if err != nil {
		return "", fmt.Errorf("failed to call greet: %w", err)
	}
	
	ptrLen := greetResults[0]
	resPtr := uint32(ptrLen >> 32)
	resLen := uint32(ptrLen)

	// 5. Ler o resultado da memória do Guest
	resBytes, ok := mod.Memory().Read(resPtr, resLen)
	if !ok {
		return "", fmt.Errorf("out of memory bounds when reading result")
	}

	return string(resBytes), nil
}

// Close libera os recursos do runtime.
func (r *Runner) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}
