# Release Guide

This guide explains how to create releases for Sidekick with auto-update support.

## Prerequisites

1. **Install GoReleaser:**
   ```bash
   # On macOS
   brew install goreleaser
   
   # On Linux
   go install github.com/goreleaser/goreleaser@latest
   ```

2. **Set up GitHub Token:**
   ```bash
   export GITHUB_TOKEN="your_github_personal_access_token"
   ```
   
   Create a token at: https://github.com/settings/tokens
   - Select: `repo` scope for private repos or `public_repo` for public

## Creating a Release

### 1. Update Version

Make sure your changes are committed:
```bash
git add .
git commit -m "feat: your changes"
git push
```

### 2. Create and Push a Tag

```bash
# Create a new tag (follow semantic versioning)
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag
git push origin v0.1.0
```

### 3. Run GoReleaser

```bash
# Create release from tag
make release

# Or run goreleaser directly
goreleaser release --clean
```

This will:
- Build binaries for Linux, macOS, and Windows (amd64 + arm64)
- Generate checksums
- Create GitHub release
- Upload all artifacts

### 4. Test Auto-Update

After the release is published:

```bash
# Check for updates (interactive mode)
sidekick
# Select: Check for Updates

# Or via CLI
sidekick update
```

## Development Releases

For testing without creating a GitHub release:

```bash
# Create snapshot build (no tags required)
make release-snapshot

# Binaries will be in ./dist/
```

## Version Information

Version is embedded at build time via ldflags:

```bash
# Manual build with version
go build -ldflags="-X github.com/pefman/sidekick/internal/updater.Version=v0.1.0" -o sidekick .

# Using Makefile (auto-detects version from git)
make build
```

## Release Checklist

- [ ] All tests pass: `make test`
- [ ] Code is formatted: `make fmt`
- [ ] No lint errors: `make lint`
- [ ] Version bumped appropriately (major.minor.patch)
- [ ] CHANGELOG.md updated
- [ ] Tag created and pushed
- [ ] Release created on GitHub
- [ ] Auto-update tested

## Semantic Versioning

Follow [semver.org](https://semver.org/):

- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v0.2.0): New features (backward compatible)
- **PATCH** (v0.1.1): Bug fixes (backward compatible)

## Troubleshooting

**Problem:** "executable is not writable" during update
- **Solution:** Run with sudo: `sudo sidekick update`

**Problem:** GoReleaser fails with "tag not found"
- **Solution:** Make sure you've pushed the tag: `git push origin v0.1.0`

**Problem:** Update not detected
- **Solution:** Check that GitHub release is published (not draft)

## CI/CD Integration

To automate releases with GitHub Actions, add `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Manual Distribution

Binaries are available at:
```
https://github.com/pefman/sidekick/releases
```

Users can download and install manually:
```bash
# Download latest release
wget https://github.com/pefman/sidekick/releases/download/v0.1.0/sidekick_Linux_x86_64.tar.gz

# Extract
tar xzf sidekick_Linux_x86_64.tar.gz

# Install
sudo mv sidekick /usr/local/sbin/
```
