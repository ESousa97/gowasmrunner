# Architecture and Design Decisions (ADR)

This document summarizes the core architectural choices of the `gowasmrunner` project.

## 1. Overview

`gowasmrunner` was designed as a hybrid orchestrator (CLI and HTTP Server) focused on the **isolated execution** of WebAssembly binary code.

The ecosystem is split into three main layers:
1. **Entry Interface (`cmd/runner`)**: Handles command interpretation, flag parsing, spinning up the web server, and dependency injection.
2. **Wasm Engine (`internal/engine`)**: The pure business abstraction. Decoupled from how the request arrives (terminal or network), it focuses entirely on memory lifecycles, function calls, binary compilation caching, and security.
3. **Wasm Plugins (`/plugins`)**: The untrusted code (Guest).

## 2. Technical Decisions

### 2.1 wazero over CGO
**Decision:** Use `github.com/tetratelabs/wazero` instead of traditional C/C++ based libraries (such as Wasmer or Wasmtime).
**Justification:** CGO libraries compromise the portability and security of Go's cross-compilation. `wazero` is 100% native to Go, resulting in a single static binary, which drastically simplifies deployment via Docker or direct downloads, while being extremely performant.

### 2.2 Linear Memory and String Management
**Context:** WebAssembly natively supports only integers and floating points. There is no native concept of a `string`.
**Adopted Solution:** 
1. The Go host calls an allocation function in the Wasm Guest (`allocate(uint64)`).
2. The Guest allocates memory and returns the offset.
3. Go writes the string as bytes directly into the linear memory slice via the wazero API.
4. The Guest processes it, writes the output to another memory location, and returns a 64-bit bitpack (`(offset << 32) | length`).
5. The Host decodes the pointer and retrieves the final string.

### 2.3 Sandboxing and Security
Running arbitrary code demands heavy restrictions.
- **MaxMemoryPages:** We limit this to 20 pages (64KB per page = 1.2MB), preventing *Out of Memory (OOM)* attacks.
- **Context Timeouts:** Every execution is tracked via `context.WithTimeout`. If a Wasm function enters an *infinite loop*, the Go scheduler issues a cancellation to the wazero runtime and halts the thread immediately.
- **Restricted WASI:** Only basic logging support to `stdout` has been enabled. The Guest **does not have** access to the local file system or network.

### 2.4 Pre-warming (Plugin Cache)
Compiling Wasm (`CompileModule`) is an expensive operation (CPU/Time). Instantiating an already compiled module is virtually instantaneous (pure memory operation).
`gowasmrunner` scans the `/plugins` directory upon initialization and stores `CompiledModule` structs in a RAM map. The Web server reuses these pre-warmed structures to serve massive and concurrent requests with zero compilation latency.
