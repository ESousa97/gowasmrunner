package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/esousa97/gowasmrunner/internal/engine"
)

func main() {
	// Define command line flags.
	wasmPath := flag.String("wasm", "", "Path to the WebAssembly module file")
	mode := flag.String("mode", "numeric", "Execution mode: 'numeric', 'string' or 'plugin'")
	pluginDir := flag.String("plugins", "./plugins", "Directory to scan for plugins")
	funcName := flag.String("func", "add", "The name of the exported function to invoke")
	flag.Parse()

	if *wasmPath == "" && *mode != "plugin" {
		fmt.Println("Usage:")
		fmt.Println("  Numeric/String: gowasmrunner -wasm <path> [-mode string|numeric] [args...]")
		fmt.Println("  Plugin:         gowasmrunner -mode plugin <plugin_name> <func_name> [args...]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()

	// Configuração do Runner com limites de segurança.
	cfg := engine.RunnerConfig{
		MaxMemoryPages: 20,                  // Aumentado para suportar plugins maiores
		Timeout:        5 * time.Second,     // Timeout de 5 segundos
		Stdout:         os.Stdout,
	}

	// Inicializar o Runner.
	runner, err := engine.NewRunner(ctx, cfg)
	if err != nil {
		log.Fatalf("Error creating runner: %v", err)
	}
	defer runner.Close(ctx)

	if *mode == "plugin" {
		// Carrega todos os plugins do diretório
		if err := runner.LoadPlugins(ctx, *pluginDir); err != nil {
			log.Fatalf("Error loading plugins: %v", err)
		}

		args := flag.Args()
		if len(args) < 2 {
			log.Fatal("Usage for plugin mode: gowasmrunner -mode plugin <plugin_file.wasm> <func_name> [args...]")
		}

		pluginFile := args[0]
		targetFunc := args[1]

		// Converte argumentos numéricos se houver
		var uintParams []uint64
		for _, arg := range args[2:] {
			val, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				log.Fatalf("Invalid argument %q: must be a positive integer", arg)
			}
			uintParams = append(uintParams, val)
		}

		results, err := runner.RunPlugin(ctx, pluginFile, targetFunc, uintParams...)
		if err != nil {
			log.Fatalf("Error executing plugin: %v", err)
		}
		fmt.Printf("Plugin %s [%s] result: %v\n", pluginFile, targetFunc, results)
		return
	}

	if *mode == "string" {

		// String Mode: get the first argument after flags.
		args := flag.Args()
		if len(args) < 1 {
			log.Fatal("Missing name argument for string mode")
		}
		
		result, err := runner.RunGreet(ctx, *wasmPath, args[0])
		if err != nil {
			log.Fatalf("Error executing wasm greeting: %v", err)
		}
		fmt.Printf("Wasm Result: %s\n", result)
	} else {
		// Numeric Mode (Original)
		args := flag.Args()
		var uintParams []uint64
		for _, arg := range args {
			val, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				log.Fatalf("Invalid argument %q: must be a positive integer", arg)
			}
			uintParams = append(uintParams, val)
		}

		results, err := runner.RunFunction(ctx, *wasmPath, *funcName, uintParams...)
		if err != nil {
			log.Fatalf("Error executing wasm function: %v", err)
		}
		fmt.Printf("Result(s) of %s(%v): %v\n", *funcName, args, results)
	}
}
