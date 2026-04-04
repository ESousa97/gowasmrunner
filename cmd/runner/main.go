package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/esousa97/gowasmrunner/internal/engine"
)

func main() {
	// Define command line flags.
	wasmPath := flag.String("wasm", "", "Path to the WebAssembly module file")
	mode := flag.String("mode", "numeric", "Execution mode: 'numeric', 'string', 'plugin' or 'server'")
	pluginDir := flag.String("plugins", "./plugins", "Directory to scan for plugins")
	funcName := flag.String("func", "add", "The name of the exported function to invoke")
	port := flag.String("port", "8080", "Port for the HTTP server to listen on")
	flag.Parse()

	if *wasmPath == "" && *mode != "plugin" && *mode != "server" {
		fmt.Println("Usage:")
		fmt.Println("  Numeric/String: gowasmrunner -wasm <path> [-mode string|numeric] [args...]")
		fmt.Println("  Plugin:         gowasmrunner -mode plugin <plugin_name> <func_name> [args...]")
		fmt.Println("  Server:         gowasmrunner -mode server [-port 8080] [-plugins ./plugins]")
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

	if *mode == "server" {
		if err := runner.LoadPlugins(ctx, *pluginDir); err != nil {
			log.Fatalf("Error loading plugins: %v", err)
		}

		http.HandleFunc("/execute/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method Not Allowed - Use POST", http.StatusMethodNotAllowed)
				return
			}

			pluginName := strings.TrimPrefix(r.URL.Path, "/execute/")
			if pluginName == "" {
				http.Error(w, "Plugin name is required in path", http.StatusBadRequest)
				return
			}

			targetFunc := r.URL.Query().Get("func")
			if targetFunc == "" {
				targetFunc = "greet" // Default para facilitar testes
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			result, err := runner.RunPluginString(r.Context(), pluginName, targetFunc, string(body))
			if err != nil {
				log.Printf("Error executing plugin %s: %v", pluginName, err)
				http.Error(w, fmt.Sprintf("Error executing plugin: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(result))
		})

		log.Printf("🚀 gowasmrunner HTTP serverless gateway running on port %s", *port)
		log.Printf("Plugins loaded from: %s", *pluginDir)
		if err := http.ListenAndServe(":"+*port, nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
		return
	}

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
