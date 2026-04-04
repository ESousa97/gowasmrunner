package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/esousa97/gowasmrunner/internal/engine"
)

func main() {
	// Definindo flags de linha de comando.
	wasmPath := flag.String("wasm", "", "Path to the WebAssembly module file")
	mode := flag.String("mode", "numeric", "Execution mode: 'numeric' or 'string'")
	flag.Parse()

	if *wasmPath == "" {
		fmt.Println("Usage: gowasmrunner -wasm <path> [-mode string|numeric] [-func <name>] [args...]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()

	// Inicializar o Runner.
	runner, err := engine.NewRunner(ctx)
	if err != nil {
		log.Fatalf("Error creating runner: %v", err)
	}
	defer runner.Close(ctx)

	if *mode == "string" {
		// Modo String: pega o primeiro argumento após as flags.
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
		// Modo Numérico (Original)
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
