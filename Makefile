.PHONY: build test test-race lint run clean tidy

BINARY_NAME := env-validator-mcp
BINARY_PATH := bin/$(BINARY_NAME)
CMD_PATH     := ./cmd/server/

## build: Compile the MCP server binary
build:
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(CMD_PATH)
	@echo "✅  Built $(BINARY_PATH)"

## test: Run all unit tests
test:
	go test ./...

## test-race: Run all tests with the race detector
test-race:
	go test -race -count=1 ./...

## coverage: Generate an HTML coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅  Coverage report: coverage.html"

## lint: Run golangci-lint (install via: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

## tidy: Tidy go.mod and go.sum
tidy:
	go mod tidy

## run: Build and run the MCP server (for manual testing)
run: build
	./$(BINARY_PATH)

## clean: Remove compiled binaries and coverage files
clean:
	rm -rf bin/ coverage.out coverage.html
