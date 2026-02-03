# Sidekick 🤖

A fast, local AI code assistant for scanning, refactoring, and planning changes across your codebase.

## Highlights
- ✅ **Prompt-first UI**: type your request immediately on launch
- ✅ **Modes**: Ask / Edit / Plan (Tab to switch)
- ✅ **Local LLM** via Ollama
- ✅ **Multi-language** and **all files** scanning
- ✅ **HTML reports** and CLI automation

## Prerequisites
1. **Install Ollama**
   ```bash
   curl -fsSL https://ollama.ai/install.sh | sh
   ```
2. **Pull a model**
   ```bash
   ollama pull qwen2.5-coder:14b-instruct-q4
   ```

## Install
```bash
# Build
cd /path/to/sidekick
make build
# or
./build.sh
# or
go build -o sidekick

# Install system-wide (optional)
sudo ./sidekick install
```

## Interactive Mode (Recommended)
```bash
sidekick
```

### How it works
- **Prompt line**: type your request immediately
- **Mode**: press **Tab** to switch Ask/Edit/Plan
- **Menu**: use **↑/↓** to select, **Enter** to open

### Modes
- **Ask**: answer questions about the code
- **Edit**: return diffs for code changes
- **Plan**: provide a step-by-step plan

## CLI Mode
```bash
# Scan a directory
sidekick scan /path/to/project

# Use a specific model
sidekick scan --model qwen2.5-coder:14b-instruct-q4

# HTML report
sidekick scan --format html --output report.html
```

## Configuration
Settings are stored at `~/.sidekick/config.json`.

Common settings:
- default model
- Ollama URL
- output format

## Supported Files
Sidekick scans **all files** (excluding hidden directories and sensitive files such as `.env`, private keys, etc.).

## Troubleshooting
- **Ollama not running**: `ollama serve`
- **Model missing**: `ollama pull <model>`

## License
MIT
