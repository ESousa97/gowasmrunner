package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// RunnerConfig define os limites de segurança para a execução.
type RunnerConfig struct {
	MaxMemoryPages uint32        // Maximum 64KB pages
	Timeout        time.Duration // Maximum execution time
	Stdout         io.Writer     // Where Wasm will write logs
}

// Runner encapsula a lógica de execução segura e o cache de plugins.
type Runner struct {
	runtime wazero.Runtime
	config  RunnerConfig
	plugins map[string]wazero.CompiledModule // Cache de módulos compilados
}

// NewRunner inicializa o runtime e o store de plugins.
func NewRunner(ctx context.Context, cfg RunnerConfig) (*Runner, error) {
	rtCfg := wazero.NewRuntimeConfig().
		WithMemoryLimitPages(cfg.MaxMemoryPages)

	r := wazero.NewRuntimeWithConfig(ctx, rtCfg)

	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &Runner{
		runtime: r,
		config:  cfg,
		plugins: make(map[string]wazero.CompiledModule),
	}, nil
}

// LoadPlugins escaneia um diretório e compila todos os arquivos .wasm encontrados.
func (r *Runner) LoadPlugins(ctx context.Context, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".wasm" {
			name := entry.Name()
			path := filepath.Join(dirPath, name)

			wasmBytes, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read plugin %s: %w", name, err)
			}

			// Compila o módulo e armazena no cache (não instancia ainda)
			compiled, err := r.runtime.CompileModule(ctx, wasmBytes)
			if err != nil {
				return fmt.Errorf("failed to compile plugin %s: %w", name, err)
			}

			r.plugins[name] = compiled
		}
	}
	return nil
}

// RunPlugin executa uma função de um plugin específico usando o módulo em cache.
func (r *Runner) RunPlugin(ctx context.Context, pluginName string, funcName string, params ...uint64) ([]uint64, error) {
	compiled, ok := r.plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found in cache", pluginName)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	modCfg := wazero.NewModuleConfig().
		WithStdout(r.config.Stdout).
		WithStderr(r.config.Stdout)

	// Instancia o módulo a partir do código já compilado (muito mais rápido)
	mod, err := r.runtime.InstantiateModule(execCtx, compiled, modCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate plugin %s: %w", pluginName, err)
	}
	defer mod.Close(execCtx)

	f := mod.ExportedFunction(funcName)
	if f == nil {
		return nil, fmt.Errorf("function %s not found in plugin %s", funcName, pluginName)
	}

	return f.Call(execCtx, params...)
}

// RunFunction carrega um arquivo Wasm e executa uma função numérica genérica.
func (r *Runner) RunFunction(ctx context.Context, wasmPath string, funcName string, params ...uint64) ([]uint64, error) {
	execCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}

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

// RunGreet carrega o módulo e executa a lógica de saudação com strings.
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

	greetFunc := mod.ExportedFunction("greet")
	allocFunc := mod.ExportedFunction("allocate")
	if greetFunc == nil || allocFunc == nil {
		return "", fmt.Errorf("exported functions 'greet' or 'allocate' not found")
	}

	nameSize := uint64(len(name))
	results, err := allocFunc.Call(execCtx, nameSize)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}
	namePtr := results[0]

	if !mod.Memory().Write(uint32(namePtr), []byte(name)) {
		return "", fmt.Errorf("out of memory bounds when writing string")
	}

	greetResults, err := greetFunc.Call(execCtx, namePtr, nameSize)
	if err != nil {
		return "", fmt.Errorf("failed to call greet: %w", err)
	}

	ptrLen := greetResults[0]
	resPtr := uint32(ptrLen >> 32)
	resLen := uint32(ptrLen)

	resBytes, ok := mod.Memory().Read(resPtr, resLen)
	if !ok {
		return "", fmt.Errorf("out of memory bounds when reading result")
	}

	return string(resBytes), nil
}

// Close libera os recursos e o cache.
func (r *Runner) Close(ctx context.Context) error {
	for _, p := range r.plugins {
		p.Close(ctx)
	}
	return r.runtime.Close(ctx)
}
