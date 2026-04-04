# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Compile the static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gowasmrunner cmd/runner/main.go

# Generate example plugins
RUN mkdir -p /app/plugins
RUN go build -o gen_greet examples/gen_greet/main.go && ./gen_greet && mv examples/greet.wasm /app/plugins/plugin-greet.wasm
RUN go build -o gen_add examples/gen_add/main.go && ./gen_add && mv examples/add.wasm /app/plugins/plugin-add.wasm

# Run Stage
FROM alpine:3.19.1

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/gowasmrunner .

# Copy the compiled plugins
COPY --from=builder /app/plugins ./plugins

EXPOSE 8080

# Start the HTTP server by default
ENTRYPOINT ["./gowasmrunner", "-mode", "server", "-port", "8080", "-plugins", "./plugins"]
