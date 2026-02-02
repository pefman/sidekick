.PHONY: build install clean test run help release release-snapshot

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/pefman/sidekick/internal/updater.Version=$(VERSION)

# Build the binary
build:
	@echo "🔨 Building sidekick $(VERSION)..."
	@go build -o sidekick -ldflags="$(LDFLAGS)" .
	@echo "✅ Build complete!"

# Install to /usr/local/sbin (requires sudo)
install: build
	@echo "📦 Installing sidekick..."
	@sudo ./sidekick install

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -f sidekick
	@echo "✅ Clean complete!"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./... -v

# Run the tool on the examples directory
demo: build
	@echo "🚀 Running demo scan on examples/..."
	@./sidekick scan examples/

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run || go vet ./...

# Create a release with goreleaser
release:
	@echo "📦 Creating release $(VERSION)..."
	@goreleaser release --clean

# Create a snapshot release (no tags required)
release-snapshot:
	@echo "📦 Creating snapshot release..."
	@goreleaser release --snapshot --clean

# Display help
help:
	@echo "Sidekick Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build             - Build the binary"
	@echo "  install           - Install to /usr/local/sbin (requires sudo)"
	@echo "  clean             - Remove build artifacts"
	@echo "  test              - Run tests"
	@echo "  demo              - Run demo scan on examples/"
	@echo "  deps              - Download and tidy dependencies"
	@echo "  fmt               - Format code"
	@echo "  lint              - Run linter"
	@echo "  release           - Create a release with goreleaser"
	@echo "  release-snapshot  - Create a snapshot release (no tags)"
	@echo "  help              - Show this help message"

# Default target
.DEFAULT_GOAL := help
