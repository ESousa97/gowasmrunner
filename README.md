# gowasmrunner

A lightweight WebAssembly module runner written in Go using the [wazero](https://github.com/tetratelabs/wazero) library.

## 🚀 Features

- **Zero CGO Dependencies:** Uses `wazero`, a 100% Go runtime.
- **WASI Support:** Ready for modules that interact with the system.
- **Flexible CLI:** Invoke exported functions by passing arguments via command line.
- **Error Handling:** Strict validation of Wasm files and function signatures.

## 🛠️ Usage

### Prerequisites
- Go 1.21+

### Installation
```bash
git clone https://github.com/esousa97/gowasmrunner.git
cd gowasmrunner
go mod download
```

### Running an example
First, generate the example module:
```bash
go run examples/gen_wasm.go
```

Now, execute the `add` function of the generated module:
```bash
go run cmd/runner/main.go -wasm examples/add.wasm -func add 10 20
```

## 🏗️ Project Structure

- `cmd/runner/`: Entry point for the CLI application.
- `internal/engine/`: Core logic for Wasm loading and execution.
- `examples/`: Example modules and generator scripts.

## 📜 License

Distributed under the MIT license. See `LICENSE` for more information.
