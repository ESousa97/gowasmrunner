# gowasmrunner

Um executor leve de módulos WebAssembly escrito em Go utilizando a biblioteca [wazero](https://github.com/tetratelabs/wazero).

## 🚀 Funcionalidades

- **Zero Dependências CGO:** Utiliza `wazero`, um runtime 100% Go.
- **Suporte WASI:** Preparado para módulos que interagem com o sistema.
- **CLI Flexível:** Invoque funções exportadas passando argumentos via linha de comando.
- **Tratamento de Erros:** Validação rigorosa de arquivos Wasm e assinaturas de função.

## 🛠️ Como Usar

### Pré-requisitos
- Go 1.21+

### Instalação
```bash
git clone https://github.com/esousa97/gowasmrunner.git
cd gowasmrunner
go mod download
```

### Executando um exemplo
Primeiro, gere o módulo de exemplo:
```bash
go run examples/gen_wasm.go
```

Agora, execute a função `add` do módulo gerado:
```bash
go run cmd/runner/main.go -wasm examples/add.wasm -func add 10 20
```

## 🏗️ Estrutura do Projeto

- `cmd/runner/`: Ponto de entrada da aplicação CLI.
- `internal/engine/`: Core logic para carregamento e execução Wasm.
- `examples/`: Módulos de exemplo e scripts de geração.

## 📜 Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.
