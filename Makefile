.PHONY: all build clean test deps lint fmt install build-all help

APP_NAME = firescan
VERSION = $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DIR = ./bin
LDFLAGS = -ldflags "-s -w"

all: deps lint test build

deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

build: deps
	@echo "🔨 Building $(APP_NAME)..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/firescan/
	@echo "✅ Built: $(BUILD_DIR)/$(APP_NAME)"
	@echo "   Run: $(BUILD_DIR)/$(APP_NAME) --help"

build-all: deps
	@echo "🔨 Cross-compiling for all platforms..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/firescan/
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/firescan/
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/firescan/
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/firescan/
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/firescan/
	@echo "✅ Cross-compilation complete"

test: deps
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "✅ Tests passed"

lint: deps
	@echo "🔍 Linting..."
	which golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./... --timeout=5m
	@echo "✅ Lint passed"

fmt:
	@echo "📝 Formatting..."
	go fmt ./...
	@echo "✅ Formatted"

clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR) coverage.txt coverage.out /tmp/firescan-*
	@echo "✅ Cleaned"

install: build
	@echo "📥 Installing to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "✅ Installed: /usr/local/bin/$(APP_NAME)"
	$(APP_NAME) --version

run: build
	@echo "🏃 Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME) $(ARGS)

help:
	@echo "🎯 FireScan Makefile"
	@echo "════════════════════"
	@echo "make          - Build for current platform"
	@echo "make deps     - Install dependencies"
	@echo "make build    - Build for current platform"
	@echo "make build-all - Cross-compile for all platforms"
	@echo "make test     - Run tests"
	@echo "make lint     - Run linter"
	@echo "make fmt      - Format code"
	@echo "make clean    - Clean build artifacts"
	@echo "make install  - Build and install to /usr/local/bin"
	@echo "make run ARGS='--help' - Build and run with args"
	@echo ""
	@echo "📦 Binary locations:"
	@echo "   $(BUILD_DIR)/$(APP_NAME)"