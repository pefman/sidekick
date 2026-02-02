# Sidekick 🤖

A CLI tool that uses local LLM (via Ollama) to scan your codebase for security vulnerabilities.

## Features

- 🎮 **Interactive Mode**: Beautiful menu-driven interface with keyboard shortcuts
- 🔍 **Code Analysis**: Analyze code with custom prompts or security scanning
- 🤖 **Local LLM**: Uses Ollama for privacy-preserving AI analysis
- ⚙️ **Persistent Settings**: Save your preferences (default model, output format, etc.)
- 🎯 **Multi-language**: Supports Go, JavaScript, TypeScript, Python, Java, C/C++, Rust, Ruby, PHP
- 📦 **Easy Installation**: Self-installs to `/usr/local/sbin` for system-wide access
- 📊 **HTML Reports**: Generate beautiful, shareable HTML reports
- 🎨 **Beautiful Output**: Color-coded severity levels with actionable recommendations

## Prerequisites

1. **Ollama**: Install from [ollama.ai](https://ollama.ai)
2. **LLM Model**: Pull a code-focused model (recommended: `codellama`)

```bash
# Install Ollama (macOS/Linux)
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull codellama
# or
ollama pull deepseek-coder
```

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/pefman/sidekick.git
cd sidekick

# Build the binary
go build -o sidekick

# Install to /usr/local/sbin (requires sudo)
sudo ./sidekick install
```

## Usage

### Interactive Mode (Recommended)

Launch the interactive menu-driven interface:

```bash
sidekick
```

This gives you a beautiful interface with:
- **Press S**: Scan your codebase
- **Press T**: Configure settings (default model, output format, etc.)
- **Press M**: View available models
- **Press H**: Show help
- **Press Q**: Quit

Your settings are saved to `~/.sidekick/config.json` and persist between sessions.

### Command Line Mode

For automation and scripts:

```bash
# Scan current directory
sidekick scan

# Scan specific path
sidekick scan /path/to/project

# Use different model
sidekick scan --model deepseek-coder

# Generate HTML report
sidekick scan --format html

# Verbose output
sidekick scan --verbose
```

## Commands

### Interactive Mode

```bash
sidekick
```

Launches the interactive interface with menu navigation and persistent settings.

### `scan`

Scans your codebase for security vulnerabilities.

**Flags:**
- `-m, --model string`: Ollama model to use (default: from config or "qwen2.5-coder:14b")
- `-v, --verbose`: Enable verbose output
- `-f, --format string`: Output format: text or html (default: from config or "text")
- `-o, --output string`: Output file for HTML reports (default: auto-generated)

**Examples:**
```bash
sidekick scan
sidekick scan ./src
sidekick scan --model deepseek-coder --verbose
sidekick scan --format html --output my-report.html
```

### `install`

Installs the sidekick binary to `/usr/local/sbin` for system-wide access.

```bash
sudo ./sidekick install
```

After installation, you can run `sidekick` from anywhere without specifying the path.

## Security Issues Detected

Sidekick scans for common vulnerabilities including:

- SQL injection risks
- Cross-site scripting (XSS) vulnerabilities
- Command injection risks
- Path traversal vulnerabilities
- Hardcoded credentials or secrets
- Insecure cryptographic practices
- Authentication/authorization issues
- Unsafe deserialization
- Resource leaks
- Race conditions
- Buffer overflows

## Example Output

```
🔍 Scanning: /home/user/project
🤖 Using model: codellama

📁 Found 15 files to analyze

📄 api/auth.go
   🔴 [CRITICAL] Hardcoded database credentials
      Line: 23
      💡 Use environment variables or a secure config management system

📄 api/user.go
   🟠 [HIGH] SQL query vulnerable to injection
      Line: 45
      💡 Use parameterized queries or prepared statements

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Scan Summary
   Total Issues: 2
   🔴 Critical: 1
   🟠 High: 1
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Supported File Types

- Go (`.go`)
- JavaScript (`.js`)
- TypeScript (`.ts`)
- Python (`.py`)
- Java (`.java`)
- C/C++ (`.c`, `.cpp`)
- Rust (`.rs`)
- Ruby (`.rb`)
- PHP (`.php`)

## Configuration

The tool automatically:
- Skips hidden directories (`.git`, `.vscode`, etc.)
- Ignores common build/dependency folders (`node_modules`, `vendor`, `dist`, `build`)
- Limits analysis to files under 100KB for performance

## Development

### Project Structure

```
sidekick/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── scan.go            # Scan command
│   └── install.go         # Install command
├── internal/
│   ├── ollama/            # Ollama client
│   │   └── client.go
│   └── scanner/           # Code analyzer
│       └── scanner.go
├── main.go                # Entry point
├── go.mod                 # Go module definition
└── README.md              # This file
```

### Building

```bash
go build -o sidekick
```

### Testing with Ollama

Make sure Ollama is running:
```bash
ollama serve
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License

## Author

Created by pefman
