# Talos Makefile

.PHONY: all build test test-e2e clean run

BINARY_NAME=talos
BUILD_DIR=bin

all: test build

build:
	@echo "🔨 Building Talos..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/atlas
	@go build -o $(BUILD_DIR)/talos-cli ./cmd/talos-cli
	@echo "✅ Build complete."

test:
	@echo "🧪 Running Unit Tests..."
	@go test -v ./internal/... ./cmd/...

test-e2e:
	@echo "🚀 Running End-to-End Tests..."
	@go test -v ./tests/e2e/...

run: build
	@echo "🔥 Starting Talos..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -rf tests/e2e/tmp
	@echo "✅ Clean documentation."
