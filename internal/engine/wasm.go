package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// RunnerConfig defines the security limits for execution.
type RunnerConfig struct {
	MaxMemoryPages uint32        // Maximum 64KB pages
	Timeout        time.Duration // Maximum execution time
	Stdout         io.Writer     // Where Wasm will write logs
}

// Runner encapsulates the secure execution logic.
type Runner struct {
	runtime wazero.Runtime
	config  RunnerConfig
}

// NewRunner initializes the runtime with global limits.
func NewRunner(ctx context.Context, cfg RunnerConfig) (*Runner, error) {
	// 1. Configure Runtime with memory limits if specified
	rtCfg := wazero.NewRuntimeConfig().
		WithMemoryLimitPages(cfg.MaxMemoryPages)

	r := wazero.NewRuntimeWithConfig(ctx, rtCfg)

	// 2. Instantiate WASI
	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &Runner{
		runtime: r,
		config:  cfg,
	}, nil
}

// RunFunction loads a Wasm file and executes a generic numeric function.
func (r *Runner) RunFunction(ctx context.Context, wasmPath string, funcName string, params ...uint64) ([]uint64, error) {
	execCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}

	// 3. Configure module with Stdout redirection
	modCfg := wazero.NewModuleConfig().
		WithStdout(r.config.Stdout).
		WithStderr(r.config.Stdout)

	mod, err := r.runtime.InstantiateWithConfig(execCtx, wasmBytes, modCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate wasm module: %w", err)
	}
	defer mod.Close(execCtx)

	f := mod.ExportedFunction(funcName)
	if f == nil {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	return f.Call(execCtx, params...)
}

// RunGreet loads the module and executes the greeting logic with strings.
func (r *Runner) RunGreet(ctx context.Context, wasmPath string, name string) (string, error) {
	execCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return "", fmt.Errorf("failed to read wasm file: %w", err)
	}

	modCfg := wazero.NewModuleConfig().
		WithStdout(r.config.Stdout).
		WithStderr(r.config.Stdout)

	mod, err := r.runtime.InstantiateWithConfig(execCtx, wasmBytes, modCfg)
	if err != nil {
		return "", fmt.Errorf("failed to instantiate wasm module: %w", err)
	}
	defer mod.Close(execCtx)

	// 1. Get exported functions
	greetFunc := mod.ExportedFunction("greet")
	allocFunc := mod.ExportedFunction("allocate")
	if greetFunc == nil || allocFunc == nil {
		return "", fmt.Errorf("exported functions 'greet' or 'allocate' not found")
	}

	// 2. Allocate memory in the Guest for the input string
	nameSize := uint64(len(name))
	results, err := allocFunc.Call(execCtx, nameSize)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}
	namePtr := results[0]

	// 3. Write the string into the Guest's linear memory
	if !mod.Memory().Write(uint32(namePtr), []byte(name)) {
		return "", fmt.Errorf("out of memory bounds when writing string")
	}

	// 4. Call the greet(ptr, len) function
	greetResults, err := greetFunc.Call(execCtx, namePtr, nameSize)
	if err != nil {
		return "", fmt.Errorf("failed to call greet: %w", err)
	}
	
	ptrLen := greetResults[0]
	resPtr := uint32(ptrLen >> 32)
	resLen := uint32(ptrLen)

	// 5. Read the result from the Guest's memory
	resBytes, ok := mod.Memory().Read(resPtr, resLen)
	if !ok {
		return "", fmt.Errorf("out of memory bounds when reading result")
	}

	return string(resBytes), nil
}

// Close frees the runtime resources.
func (r *Runner) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}
