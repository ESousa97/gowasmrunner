# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Otimização do cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código-fonte
COPY . .

# Compila o binário estático
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gowasmrunner cmd/runner/main.go

# Gera os plugins de exemplo (Wasm manual via script Go)
RUN mkdir -p /app/plugins
RUN go run examples/gen_greet.go && mv examples/greet.wasm /app/plugins/plugin-greet.wasm
RUN go run examples/gen_wasm.go && mv examples/add.wasm /app/plugins/plugin-add.wasm

# Run Stage
FROM alpine:latest

WORKDIR /root/

# Copia o binário
COPY --from=builder /app/gowasmrunner .

# Copia os plugins compilados
COPY --from=builder /app/plugins ./plugins

EXPOSE 8080

# Inicia o servidor HTTP por padrão
ENTRYPOINT ["./gowasmrunner", "-mode", "server", "-port", "8080", "-plugins", "./plugins"]
