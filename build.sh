#!/bin/bash

# Build script for sidekick

set -e

echo "🔨 Building sidekick..."

# Build for current platform
go build -o sidekick -ldflags="-s -w" .

echo "✅ Build complete!"
echo ""
echo "To install system-wide:"
echo "  sudo ./sidekick install"
echo ""
echo "To run directly:"
echo "  ./sidekick scan"
