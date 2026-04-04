.PHONY: build test clean run gen-example

BINARY_NAME=gowasmrunner

build:
	go build -o bin/$(BINARY_NAME) cmd/runner/main.go

test:
	go test ./...

gen-example:
	go run examples/gen_wasm.go

run: build gen-example
	./bin/$(BINARY_NAME) -wasm examples/add.wasm -func add 10 20

clean:
	rm -rf bin/
	rm -f examples/add.wasm
