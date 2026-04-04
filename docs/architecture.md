# Arquitetura e Decisões de Design (ADR)

Este documento sumariza as principais escolhas arquiteturais do projeto `gowasmrunner`.

## 1. Visão Geral

O `gowasmrunner` foi projetado como um orquestrador híbrido (CLI e Servidor HTTP) focado na **execução isolada** de código binário WebAssembly.

O ecossistema é dividido em três camadas principais:
1. **Interface de Entrada (`cmd/runner`)**: Lida com a interpretação de comandos, parsing de flags, subida do servidor web e injeção de dependências.
2. **Motor Wasm (`internal/engine`)**: A abstração pura de negócio. Desacoplada de como o request chega (terminal ou rede), foca exclusivamente em ciclo de vida de memória, chamadas de funções, cache de binários compilados e segurança.
3. **Wasm Plugins (`/plugins`)**: O código não confiável (Guest).

## 2. Decisões Técnicas

### 2.1 wazero sobre CGO
**Decisão:** Utilizar `github.com/tetratelabs/wazero` em vez de bibliotecas tradicionais baseadas em C/C++ (como Wasmer ou Wasmtime).
**Justificativa:** Bibliotecas CGO comprometem a portabilidade e segurança da compilação cruzada do Go. O `wazero` é 100% nativo em Go, resultando em um único binário estático, o que simplifica drasticamente a implantação via Docker ou downloads diretos, além de ser extremamente performático.

### 2.2 Gerenciamento Linear de Memória e Strings
**Contexto:** O WebAssembly tem suporte nativo apenas para inteiros e pontos flutuantes. Não existe o conceito nativo de `string`.
**Solução Adotada:** 
1. O host Go chama uma função de alocação no Wasm Guest (`allocate(uint64)`).
2. O Guest aloca a memória e devolve o offset.
3. O Go escreve a string em bytes diretamente no slice de memória linear através da API do wazero.
4. O Guest processa, escreve a saída em outro ponto da memória e devolve um bitpack de 64 bits (`(offset << 32) | length`).
5. O Host decodifica o ponteiro e obtém a string final.

### 2.3 Sandboxing e Segurança
Executar código arbitrário demanda restrições pesadas.
- **MaxMemoryPages:** Limitamos a 20 páginas (64KB por página = 1.2MB), prevenindo *Out of Memory (OOM)*.
- **Context Timeouts:** Toda execução é rastreada por `context.WithTimeout`. Se uma função Wasm entrar num *loop infinito*, o scheduler do Go emite o cancelamento para o runtime do wazero e interrompe a thread imediatamente.
- **WASI Restrita:** Apenas o suporte básico a log no `stdout` foi habilitado. O Guest **não possui** acesso a file system local ou rede.

### 2.4 Pre-warming (Plugin Cache)
Compilar Wasm (`CompileModule`) é caro (CPU/Tempo). Instanciar um módulo já compilado é virtualmente instantâneo (memória pura).
O `gowasmrunner` escaneia o diretório `/plugins` na inicialização, e armazena structs `CompiledModule` em um mapa na RAM. O servidor Web reutiliza estas estruturas pré-aquecidas para servir requisições massivas e concorrentes sem latência de compilação.
