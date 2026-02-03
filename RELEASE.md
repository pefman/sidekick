# Release Guide

## Prerequisites
- Go installed
- GitHub token in `GITHUB_TOKEN`
- GoReleaser installed

```bash
go install github.com/goreleaser/goreleaser@latest
```

## Release Steps
1. Commit changes
2. Tag version
3. Run GoReleaser

```bash
git add .
git commit -m "release: v1.1.0"
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

goreleaser release --clean
```

## Notes
- `Version` is embedded via ldflags in build scripts.
- For testing: `goreleaser release --snapshot --clean`
