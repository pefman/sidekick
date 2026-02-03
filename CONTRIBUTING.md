# Contributing to Sidekick

Thank you for your interest in contributing to Sidekick!

## Development Setup

1. **Prerequisites**
   - Go 1.21 or later
   - Ollama installed locally
   - Git

2. **Clone and Build**
   ```bash
   git clone https://github.com/pefman/sidekick.git
   cd sidekick
   go mod download
   make build
   ```

## Project Structure

```
sidekick/
├── cmd/                   # CLI commands (cobra)
│   ├── root.go           # Root command setup
│   ├── scan.go           # Scan command
│   └── install.go        # Installation command
├── internal/
│   ├── interactive/      # Prompt-first UI
│   ├── prompts/          # Prompt templates
│   ├── ollama/           # Ollama API client
│   └── scanner/          # Scan/analysis logic
├── examples/             # Example code
├── main.go               # Entry point
└── README.md
```

## Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing code style
   - Add comments for exported functions

3. **Test your changes**
   ```bash
   make build
   ./sidekick
   ```

4. **Format and lint**
   ```bash
   make fmt
   make lint
   ```

## Future Enhancement Ideas

- [ ] Add support for custom security rules
- [ ] Implement caching for faster re-scans
- [ ] Add JSON/SARIF output formats
- [ ] Support for more programming languages
- [ ] Integration with CI/CD pipelines
- [ ] Configuration file support
- [ ] Progress bar for large scans
- [ ] Parallel file processing
- [ ] Custom prompt templates
- [ ] Severity filtering
- [ ] Exclude patterns configuration
- [ ] Report generation (HTML/PDF)
- [ ] Git hook integration
- [ ] IDE plugins

## Adding New Commands

1. Create a new file in `cmd/` (e.g., `cmd/mycommand.go`)
2. Define the command using cobra
3. Register it in `cmd/root.go`'s `init()` function

Example:
```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

## Prompts & Modes

- Ask/Edit/Plan prompts live in `internal/prompts/custom/`
- Use `internal/prompts` helpers for rendering
- Keep prompt changes documented in release notes

## Code Style

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Write descriptive variable names
- Add comments for exported functions
- Keep functions focused and small

## Testing

Currently, the project includes example files for manual testing.
Future: Add unit tests and integration tests.

```bash
# Run tests (when available)
go test ./...

# Run with coverage
go test -cover ./...
```

## Pull Request Process

1. Update README.md if needed
2. Ensure code builds and runs
3. Write clear commit messages
4. Submit PR with description of changes
5. Respond to review feedback

## Questions?

Feel free to open an issue for:
- Bug reports
- Feature requests
- Questions about the code
- Documentation improvements

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
