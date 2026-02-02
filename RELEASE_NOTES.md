# Sidekick v0.1.0 - Interactive CLI Release

## 🎉 What's New

Sidekick has been transformed into a **fully interactive CLI tool** with persistent settings and an intuitive menu-driven interface!

## 🎮 Interactive Mode

Simply run `./sidekick` (no arguments) to launch the interactive interface:

```
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║           🤖  SIDEKICK  - AI Code Security Scanner        ║
║                                                            ║
║              Your AI-Powered Security Assistant            ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━ MAIN MENU ━━━━━━━━━━━━━━━━

  1. [S]can      - Scan codebase for security issues
  2. [T]ettings  - Configure settings
  3. [M]odels    - View/manage LLM models
  4. [H]elp      - Show help information
  5. [Q]uit      - Exit Sidekick
```

### Keyboard Shortcuts

Navigate quickly with single keypresses:
- `S` or `1` - Scan
- `T` or `2` - Settings
- `M` or `3` - Models
- `H` or `4` - Help
- `Q` or `5` - Quit

## ⚙️ Persistent Settings

Your preferences are saved in `~/.sidekick/config.json`:

```json
{
  "default_model": "qwen2.5-coder:14b",
  "ollama_url": "http://localhost:11434",
  "verbose": false,
  "output_format": "text"
}
```

### Settings Menu

Press `T` in the main menu to access:

1. **Change default model** - Set your preferred LLM
2. **Change Ollama URL** - Connect to remote Ollama instance
3. **Toggle verbose mode** - Enable/disable detailed output
4. **Change default output format** - Choose between text/html
5. **Reset to defaults** - Restore default settings

## 📊 Interactive Scanning

When you choose "Scan" from the menu:

1. **Enter scan path** - Choose directory or press Enter for current
2. **Select output format** - Text (console) or HTML report
3. **Choose model** - Use default or specify different model
4. **Confirm and scan** - Review settings and start

## 🤖 Model Management

Press `M` to view all available Ollama models with the default marked:

```
✓ 1. qwen2.5-coder:14b
  2. deepseek-r1:14b
  3. deepseek-r1:32b
  4. Qwen2.5:1.5B
```

## 🚀 Command Line Mode (Still Available)

For scripts and automation, use traditional CLI commands:

```bash
# Quick scan with defaults from config
./sidekick scan

# Override settings
./sidekick scan --model deepseek-r1:14b
./sidekick scan --format html --output report.html
./sidekick scan /path/to/project --verbose

# Install system-wide
sudo ./sidekick install
```

## 📁 Project Structure

```
sidekick/
├── cmd/
│   ├── root.go        - Interactive mode entry point
│   ├── scan.go        - Scan command with config support
│   └── install.go     - Installation command
├── internal/
│   ├── config/        - Configuration management
│   ├── interactive/   - Interactive UI components
│   │   ├── menu.go    - Main menu and navigation
│   │   └── scan.go    - Interactive scanning
│   ├── ollama/        - Ollama API client
│   ├── report/        - HTML report generator
│   └── scanner/       - Security scanner
└── examples/          - Test vulnerable code
```

## 🎯 Usage Examples

### Interactive Workflow

```bash
# Launch interactive mode
./sidekick

# Press 'T' to enter settings
# Choose option 1 to change default model
# Enter: deepseek-r1:14b

# Press 'S' to scan
# Enter path: ./examples
# Choose HTML output
# Use default model
# Confirm and scan
```

### Command Line Workflow

```bash
# First time: set up config via interactive mode
./sidekick
# Press T → Configure settings → Q to quit

# Now use CLI with your saved settings
./sidekick scan                    # Uses your defaults
./sidekick scan --format html      # Override format
./sidekick scan -m Qwen2.5:1.5B   # Override model
```

## 🆕 New Features Summary

1. ✅ **Interactive menu-driven interface**
2. ✅ **Persistent configuration** (~/.sidekick/config.json)
3. ✅ **Settings management** (default model, format, etc.)
4. ✅ **Model browser** (view all available models)
5. ✅ **Interactive scan wizard** (step-by-step prompts)
6. ✅ **Keyboard shortcuts** (S/T/M/H/Q navigation)
7. ✅ **Config-aware CLI** (commands use saved defaults)
8. ✅ **Beautiful UI** (colors, borders, clear navigation)

## 🔄 Migration from v0.0.x

If you were using command-line only:
- **No breaking changes!** All CLI commands work as before
- **New**: Run without arguments for interactive mode
- **New**: Settings persist between sessions
- **Improved**: CLI commands now use your saved defaults

## 📝 Quick Reference

```bash
# Interactive mode (recommended for humans)
./sidekick

# Command line mode (good for scripts)
./sidekick scan [options]

# View help
./sidekick --help
./sidekick scan --help

# Install system-wide
sudo ./sidekick install

# Demo
./demo.sh
```

## 🐛 Known Limitations

- Settings file is user-specific (not project-specific)
- Interactive mode requires terminal with ANSI color support
- No config file location override yet

## 🚧 Future Enhancements

- Project-specific `.sidekick.yaml` config files
- Scan history and result caching
- Interactive result browser
- Export to multiple formats (JSON, SARIF, PDF)
- CI/CD integration helpers

---

**Enjoy the new interactive experience!** 🎮🚀
