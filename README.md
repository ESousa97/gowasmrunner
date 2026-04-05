<div align="center">
  <h1>Go Wasm Runner</h1>
  <p>Ultra-lightweight serverless execution environment and gateway for WebAssembly modules — CLI, HTTP API, sandboxing, and plugin hot-reload, written 100% in Go.</p>

  <img src="assets/github-go.png" alt="Go Wasm Runner Banner" width="600px">

  <br>

  ![CI](https://img.shields.io/github/actions/workflow/status/ESousa97/gowasmrunner/ci.yml?style=flat&logo=githubactions&logoColor=white)
  [![Go Report Card](https://goreportcard.com/badge/github.com/ESousa97/gowasmrunner?style=flat)](https://goreportcard.com/report/github.com/ESousa97/gowasmrunner)
  [![CodeFactor](https://www.codefactor.io/repository/github/esousa97/gowasmrunner/badge?style=flat)](https://www.codefactor.io/repository/github/esousa97/gowasmrunner)
  [![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ESousa97/gowasmrunner)
  ![License](https://img.shields.io/github/license/ESousa97/gowasmrunner?style=flat&color=blue)
  ![Go Version](https://img.shields.io/github/go-mod/go-version/ESousa97/gowasmrunner?style=flat&logo=go&logoColor=white)
  ![Last Commit](https://img.shields.io/github/last-commit/ESousa97/gowasmrunner?style=flat)
</div>

---

**gowasmrunner** is an isolated execution environment that loads and runs WebAssembly functions locally via CLI or exposes them as a serverless HTTP API. It solves portability and security challenges when executing third-party code by enforcing strict memory and execution-time boundaries, with zero CGO dependencies and a language-agnostic plugin model.

## Technologies & Frameworks

<div align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/wazero-00ADD8?style=flat&logo=go&logoColor=white" alt="wazero" />
  <img src="https://img.shields.io/badge/WebAssembly-654FF0?style=flat&logo=webassembly&logoColor=white" alt="WebAssembly" />
  <img src="https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/WASI-000000?style=flat&logo=webassembly&logoColor=white" alt="WASI" />
</div>

## Development Roadmap

- [x] **Phase 1: The Host (Basic Wasm Runtime)**
  - **Objective:** Configure the wazero runtime and execute a simple arithmetic function compiled to Wasm.
  - **Status:** CGO-free wazero runtime configured; `.wasm` file loading and exported function invocation via CLI completed.

- [x] **Phase 2: The Bridge (Memory and Data Exchange)**
  - **Objective:** Overcome Wasm's numeric-only boundary to enable string and binary payload exchange.
  - **Status:** Guest-side allocator and host-side linear memory read/write implemented; `greet` string round-trip validated.

- [x] **Phase 3: The Warden (Sandboxing and Resource Limits)**
  - **Objective:** Prevent unbounded resource consumption — essential for the serverless model.
  - **Status:** `MaxMemoryPages` and `context.WithTimeout` enforced per invocation; WASI stdout isolation activated.

- [x] **Phase 4: The Registry (Dynamic Plugin System)**
  - **Objective:** Transform the single-module executor into a multi-plugin platform with hot compilation.
  - **Status:** `PluginStore` scans `/plugins`, pre-compiles modules to `CompiledModule`, and caches them for sub-millisecond re-invocations.

- [x] **Phase 5: The Gateway (Serverless HTTP Interface)**
  - **Objective:** Expose Wasm modules through an HTTP API, mirroring AWS Lambda or Cloudflare Workers behavior.
  - **Status:** HTTP server on port 8080 with `/execute/{plugin_name}` routing; lean Alpine `Dockerfile` ready for deployment.

- [x] **Language Agnosticism (The Cherry on Top)**
  - **Objective:** Prove the runtime is language-agnostic.
  - **Status:** Rust-compiled Fibonacci and TinyGo modules run identically alongside Go-compiled plugins.

## Quick Start

### Prerequisites

- **Go 1.24+**
- **Docker** (optional, for containerized deployment)
- **golangci-lint** (optional, for local linting)

### Installation

```bash
# Install directly as a binary
go install github.com/ESousa97/gowasmrunner/cmd/runner@latest

# Or build from source
git clone https://github.com/ESousa97/gowasmrunner.git
cd gowasmrunner
make build
```

### Running

**CLI mode — numeric function:**
```bash
make run
# gowasmrunner -mode numeric -wasm examples/add.wasm -func add 10 20
# Result: [30]
```

**CLI mode — string function:**
```bash
./bin/gowasmrunner -mode string -wasm examples/greet.wasm -func greet World
# Result: Hello, World
```

**Serverless gateway:**
```bash
./bin/gowasmrunner -mode server -port 8080 -plugins ./plugins

# In another terminal:
curl -X POST "http://localhost:8080/execute/plugin-greet.wasm?func=greet" -d "World"
# Hello, World
```

**With Docker:**
```bash
docker build -t gowasmrunner .
docker run -p 8080:8080 gowasmrunner
```

### Demo

![gowasmrunner in action](assets/gowasmrunner_in_action.png)

## Makefile Targets

| Target | Description |
|---|---|
| `make build` | Compiles the CLI binary into the `bin/` folder |
| `make test` | Runs all integration and unit tests |
| `make gen-example` | Generates the `add.wasm` example plugin |
| `make gen-all` | Generates all example `.wasm` modules |
| `make run` | Builds, generates examples, and runs a numeric addition test |
| `make run-fibonacci` | Builds, generates all plugins, and runs the Rust Fibonacci test |
| `make clean` | Removes build artifacts and compiled `.wasm` modules |

## Architecture

The project follows a modular architecture centred on isolation between the host runtime and guest modules.

<div align="center">

```mermaid
graph TD
    CLI[CLI / HTTP Client] --> Runner[cmd/runner]

    subgraph "Execution Layer"
        Runner --> Engine[internal/engine]
        Engine --> Store[PluginStore Cache]
        Engine --> Limits[Memory + Timeout Limits]
        Engine --> WASI[WASI Stdout Bridge]
    end

    subgraph "Plugin Layer"
        Store --> WasmGo[Go / TinyGo Plugin]
        Store --> WasmRust[Rust Plugin]
        Store --> WasmCustom[Any .wasm Plugin]
    end

    subgraph "HTTP Gateway"
        Runner --> HTTP[net/http Server :8080]
        HTTP --> Execute[POST /execute/{plugin}]
        Execute --> Engine
    end

    style Engine fill:#2da44e,stroke:#fff,stroke-width:1px,color:#fff
    style HTTP fill:#3498db,stroke:#fff,stroke-width:1px,color:#fff
    style Store fill:#8e44ad,stroke:#fff,stroke-width:1px,color:#fff
```

</div>

- `cmd/runner`: Entry point; orchestrates CLI flags, plugin loading, and the HTTP server lifecycle.
- `internal/engine`: Core runtime. Manages the wazero lifecycle, resource limits (memory/timeout), WASI integration, and the compiled-module cache (`PluginStore`).
- `examples/`: Generator programs that produce `.wasm` binaries used by integration tests and demos.
- `plugins/`: Default directory scanned at startup for module pre-warming.

## Configuration

Current configurations are managed via command-line flags:

| Flag | Type | Default | Description |
|---|---|---|---|
| `-mode` | string | `numeric` | Operation mode: `numeric`, `string`, `plugin`, `server` |
| `-wasm` | string | `""` | Direct path to a `.wasm` file (numeric / string modes) |
| `-plugins` | string | `./plugins` | Directory scanned for cached plugin modules |
| `-func` | string | `add` | Default exported function to invoke |
| `-port` | string | `8080` | HTTP server listen port (server mode) |

## API Reference

### Execute Plugin

`POST /execute/{plugin_name}`

Executes an exported function from a pre-compiled and cached Wasm module.

| Parameter | Location | Description |
|---|---|---|
| `plugin_name` | path | Filename of the `.wasm` plugin (e.g. `plugin-greet.wasm`) |
| `func` | query | Name of the exported function to call. Default: `greet` |
| _(payload)_ | body | Raw string passed into the module's linear memory |

**Response:** Plain-text result returned by the Wasm function.

| Method | Path | Description |
|---|---|---|
| `POST` | `/execute/{plugin_name}` | Invokes an exported function from the named plugin |

## License

Distributed under the MIT License. See `LICENSE` for more information.

<div align="center">

## Author

**Enoque Sousa**

[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=flat&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/enoque-sousa-bb89aa168/)
[![GitHub](https://img.shields.io/badge/GitHub-100000?style=flat&logo=github&logoColor=white)](https://github.com/ESousa97)
[![Portfolio](https://img.shields.io/badge/Portfolio-FF5722?style=flat&logo=target&logoColor=white)](https://enoquesousa.vercel.app)

**[⬆ Back to Top](#go-wasm-runner)**

Made with ❤️ by [Enoque Sousa](https://github.com/ESousa97)

**Project Status:** Active Development

</div>
