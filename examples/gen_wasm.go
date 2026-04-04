package main

import (
	"log"
	"os"
)

func main() {
	// Binário Wasm mínimo para uma função add(i32, i32) -> i32
	// WAT:
	// (module
	//   (func $add (param i32 i32) (result i32)
	//     local.get 0
	//     local.get 1
	//     i32.add)
	//   (export "add" (func 0))
	// )
	wasmBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // Header
		0x01, 0x07, 0x01, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, // Type Section
		0x03, 0x02, 0x01, 0x00, // Function Section
		0x07, 0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, // Export Section
		0x0a, 0x09, 0x01, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b, // Code Section
	}

	err := os.WriteFile("examples/add.wasm", wasmBinary, 0644)
	if err != nil {
		log.Fatalf("failed to write wasm file: %v", err)
	}
	log.Println("Created examples/add.wasm successfully")
}
