# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-04-04

### Added
- **Serverless Gateway:** `server` mode exposing a high-performance HTTP endpoint (`/execute/{plugin}`) on port 8080.
- **Plugin Management:** Ability to scan a directory (`/plugins`) and cache compiled WebAssembly modules for instant execution.
- **Resource Limits:** Hard limits on maximum memory allocation (`MaxMemoryPages`) and execution timeouts via `context.Context` to protect the host environment.
- **String Support:** Advanced linear memory management to securely pass and return strings between Go Host and Wasm Guest.
- **WASI Support:** Standard integration for secure standard output logging from guest modules.
- **Docker Support:** Ready-to-deploy `Dockerfile` containing pre-built example plugins and a lightweight Alpine environment.
- **Testing:** Integration tests covering numerical computation, string manipulation, timeouts, and memory exhaustion.

### Fixed
- Stabilized Wasm binary generation scripts to correctly format Type and Code sections for testing environments.
