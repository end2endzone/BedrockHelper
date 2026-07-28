# Define variables
BINARY_NAME=bedrock_helper
CMD_DIR=./cmd/bedrock_helper

.PHONY: all build run test clean tidy

all: build

## build: Compile the binary
build:
	@echo "Building binary..."
	go build -o bin/$(BINARY_NAME) src/$(CMD_DIR)

## run: Build and execute the application
run: build
	@echo "Running application..."
	./bin/$(BINARY_NAME)

## test: Run unit tests across all packages
test:
	@echo "Running tests..."
	go test -v ./src/...

## clean: Remove compiled binaries
clean:
	@echo "Cleaning build cache..."
	rm -rf bin/

## tidy: Format code and clean up dependencies
tidy:
	go fmt ./src/...
	go mod tidy
