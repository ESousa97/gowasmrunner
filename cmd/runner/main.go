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
	funcName := flag.String("func", "add", "The name of the exported function to invoke")
	flag.Parse()

	if *wasmPath == "" {
		fmt.Println("Usage: gowasmrunner -wasm <path> -func <name> [arg1 arg2 ...]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Converter argumentos restantes em uint64.
	args := flag.Args()
	var uintParams []uint64
	for _, arg := range args {
		val, err := strconv.ParseUint(arg, 10, 64)
		if err != nil {
			log.Fatalf("Invalid argument %q: must be a positive integer", arg)
		}
		uintParams = append(uintParams, val)
	}

	ctx := context.Background()

	// Inicializar o Runner.
	runner, err := engine.NewRunner(ctx)
	if err != nil {
		log.Fatalf("Error creating runner: %v", err)
	}
	defer runner.Close(ctx)

	// Executar a função especificada.
	results, err := runner.RunFunction(ctx, *wasmPath, *funcName, uintParams...)
	if err != nil {
		log.Fatalf("Error executing wasm function: %v", err)
	}

	// Exibir os resultados.
	if len(results) > 0 {
		fmt.Printf("Result(s) of %s(%v): %v\n", *funcName, args, results)
	} else {
		fmt.Printf("Function %s completed with no return value\n", *funcName)
	}
}
