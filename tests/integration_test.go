package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/esousa97/gowasmrunner/internal/engine"
)

func TestWasmRunner(t *testing.T) {
	ctx := context.Background()
	
	// Inicializa o runner
	runner, err := engine.NewRunner(ctx)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}
	defer runner.Close(ctx)

	// Subteste 1: Operação Numérica (Soma)
	t.Run("Numeric Addition", func(t *testing.T) {
		wasmPath := filepath.Join("..", "examples", "add.wasm")
		
		// Verifica se o arquivo existe, senão pula (deve ser gerado antes)
		if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
			t.Skip("add.wasm not found, run generators first")
		}

		results, err := runner.RunFunction(ctx, wasmPath, "add", 10, 20)
		if err != nil {
			t.Fatalf("RunFunction failed: %v", err)
		}

		if len(results) == 0 || results[0] != 30 {
			t.Errorf("expected 30, got %v", results)
		}
	})

	// Subteste 2: Manipulação de Strings (Saudação)
	t.Run("String Greeting", func(t *testing.T) {
		wasmPath := filepath.Join("..", "examples", "greet.wasm")
		
		if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
			t.Skip("greet.wasm not found, run generators first")
		}

		name := "Gemini"
		expected := "Hello, Gemini"
		
		result, err := runner.RunGreet(ctx, wasmPath, name)
		if err != nil {
			t.Fatalf("RunGreet failed: %v", err)
		}

		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	// Subteste 3: Erro com arquivo inválido
	t.Run("Invalid Wasm File", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "invalid.wasm")
		os.WriteFile(tmpFile, []byte("not a wasm file"), 0644)

		_, err := runner.RunFunction(ctx, tmpFile, "add", 1, 2)
		if err == nil {
			t.Error("expected error for invalid wasm file, got nil")
		}
	})
}
