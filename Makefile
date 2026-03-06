TOOL_BIN = bin/gotools/$(shell uname -s)-$(shell uname -m)
BIN_OUTPUT_PATH = bin

build:
	rm -f $(BIN_OUTPUT_PATH)/generic-gantry
	go build -ldflags="-s -w" -o $(BIN_OUTPUT_PATH)/generic-gantry main.go

setup:
	go mod download

module: build
	rm -f $(BIN_OUTPUT_PATH)/module.tar.gz
	tar czf $(BIN_OUTPUT_PATH)/module.tar.gz $(BIN_OUTPUT_PATH)/generic-gantry meta.json

tool-install:
	GOBIN=`pwd`/$(TOOL_BIN) go install \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		gotest.tools/gotestsum

lint: tool-install
	go vet ./...
	GOGC=50 $(TOOL_BIN)/golangci-lint run -v --fix --config=./etc/.golangci.yaml

test:
	go test -race ./...

clean:
	rm -rf $(BIN_OUTPUT_PATH)
