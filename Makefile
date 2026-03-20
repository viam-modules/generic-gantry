BIN_OUTPUT_PATH = bin

build:
	rm -f $(BIN_OUTPUT_PATH)/generic-gantry
	go build -ldflags="-s -w" -o $(BIN_OUTPUT_PATH)/generic-gantry main.go

setup:
	go mod download

module: build
	rm -f $(BIN_OUTPUT_PATH)/module.tar.gz
	tar czf $(BIN_OUTPUT_PATH)/module.tar.gz $(BIN_OUTPUT_PATH)/generic-gantry meta.json

lint:
	go mod tidy
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run -v --fix --config=./etc/.golangci.yaml --timeout 5m

test:
	go test -race ./...

clean:
	rm -rf $(BIN_OUTPUT_PATH)
