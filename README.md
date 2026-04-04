# gowasmrunner

> Um executor e gateway serverless ultraleve para módulos WebAssembly escrito 100% em Go.

![CI](https://github.com/ESousa97/gowasmrunner/actions/workflows/ci.yml/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/ESousa97/gowasmrunner)
![Go Reference](https://pkg.go.dev/badge/github.com/ESousa97/gowasmrunner.svg)
![License](https://img.shields.io/github/license/ESousa97/gowasmrunner)
![Go Version](https://img.shields.io/github/go-mod/go-version/ESousa97/gowasmrunner)
![Last Commit](https://img.shields.io/github/last-commit/ESousa97/gowasmrunner)

---

O gowasmrunner é um ambiente de execução isolado que permite rodar funções WebAssembly localmente via CLI ou expô-las instantaneamente como uma API HTTP serverless. Ele resolve o problema de portabilidade e segurança na execução de código de terceiros, garantindo limites rígidos de memória e tempo de execução sem depender de dependências CGO.

## Demonstração

Executando um plugin via CLI:
```bash
$ gowasmrunner -mode plugin plugin-add.wasm add 10 20
Plugin plugin-add.wasm [add] result: [30]
```

Executando como Gateway Serverless:
```bash
$ gowasmrunner -mode server -port 8080
2026/04/04 19:30:00 🚀 gowasmrunner HTTP serverless gateway running on port 8080

# Em outro terminal:
$ curl -X POST "http://localhost:8080/execute/plugin-greet.wasm?func=greet" -d "Mundo"
Hello, Mundo
```

## Stack Tecnológico

| Tecnologia | Papel |
|---|---|
| Go | Linguagem base, provê concorrência e compilação estática |
| wazero | Runtime WebAssembly zero-dependências (CGO-free) |
| Docker | Empacotamento e distribuição do Gateway Serverless |
| net/http | Servidor web nativo para exposição dos plugins |

## Pré-requisitos

- Go >= 1.21
- Docker (opcional, para rodar como container)

## Instalação e Uso

### Como binário

```bash
go install github.com/ESousa97/gowasmrunner/cmd/runner@latest
```

### A partir do source

```bash
git clone https://github.com/ESousa97/gowasmrunner.git
cd gowasmrunner
make build
make run
```

### Com Docker

```bash
docker build -t gowasmrunner .
docker run -p 8080:8080 gowasmrunner
```

## Makefile Targets

| Target | Descrição |
|---|---|
| `build` | Compila o binário CLI na pasta `bin/` |
| `test` | Executa todos os testes de integração e unidade |
| `gen-example` | Gera os módulos `.wasm` de exemplo na pasta `examples/` |
| `run` | Faz o build, gera os exemplos e executa uma soma simples de teste |
| `clean` | Remove artefatos de build e módulos compilados temporários |

## Arquitetura

O projeto segue uma arquitetura modular focada em isolamento:
- `cmd/runner`: Ponto de entrada que gerencia a CLI e o Servidor HTTP.
- `internal/engine`: Core do sistema. Gerencia o ciclo de vida do `wazero`, limites de recursos (memória/timeout), integração WASI e o cache de módulos compilados (`PluginStore`).
- `plugins/`: Diretório padrão escaneado pelo sistema para pre-warming de módulos.

Veja [docs/architecture.md](docs/architecture.md) para mais detalhes sobre as decisões técnicas.

## API Reference

### Executar Plugin

`POST /execute/{plugin_name}`

Executa uma função exportada de um módulo Wasm em cache.

**Query Parameters:**
- `func` (opcional): Nome da função a ser executada. Padrão: `greet`.

**Body:**
Payload bruto (text/plain, application/json, etc) que será passado para a memória linear do módulo Wasm.

**Response:**
O resultado retornado pela função Wasm, codificado como texto no corpo da resposta.

## Configuração

As configurações atuais são geridas via flags de linha de comando:

| Flag | Descrição | Tipo | Padrão |
|---|---|---|---|
| `-mode` | Modo de operação (`numeric`, `string`, `plugin`, `server`) | string | `numeric` |
| `-wasm` | Caminho direto para um arquivo Wasm (modos num/str) | string | `""` |
| `-plugins` | Diretório para carregar plugins em cache | string | `./plugins` |
| `-func` | Função padrão a ser executada | string | `add` |
| `-port` | Porta para o servidor HTTP | string | `8080` |

## Roadmap

- [x] Suporte básico de execução Wasm
- [x] Passagem de strings via memória linear
- [x] Limites de segurança (Memory e Timeout)
- [x] Sistema de Cache de Plugins
- [x] Gateway HTTP Serverless
- [ ] Hot-reload de plugins via filesystem watcher
- [ ] Suporte a passagem estruturada de JSON (Host <-> Guest)
- [ ] Métricas de execução via Prometheus

## Contribuindo

Consulte nosso [CONTRIBUTING.md](CONTRIBUTING.md) para saber como configurar seu ambiente, rodar os testes e enviar Pull Requests.

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para mais informações.

## Autor

Enoque Sousa - [Portfólio](https://enoquesousa.vercel.app) - [GitHub](https://github.com/ESousa97)
