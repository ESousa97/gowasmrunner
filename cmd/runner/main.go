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
	mode := flag.String("mode", "numeric", "Execution mode: 'numeric' or 'string'")
	funcName := flag.String("func", "add", "The name of the exported function to invoke")
	flag.Parse()

	if *wasmPath == "" {
		fmt.Println("Usage: gowasmrunner -wasm <path> [-mode string|numeric] [-func <name>] [args...]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()

	// Runner configuration with safety limits.
	cfg := engine.RunnerConfig{
		MaxMemoryPages: 10,                 // Limit of 640KB RAM
		Timeout:        2 * time.Second,     // 2 seconds timeout
		Stdout:         os.Stdout,           // Wasm logs to terminal
	}

	// Initialize the Runner.
	runner, err := engine.NewRunner(ctx, cfg)
	if err != nil {
		log.Fatalf("Error creating runner: %v", err)
	}
	defer runner.Close(ctx)

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
