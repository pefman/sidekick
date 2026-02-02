# Quick Start Guide

## Prerequisites

1. **Install Ollama**
   ```bash
   curl -fsSL https://ollama.ai/install.sh | sh
   ```

2. **Pull a code model**
   ```bash
   ollama pull codellama
   # Alternative models:
   # ollama pull deepseek-coder
   # ollama pull llama2
   ```

3. **Start Ollama service**
   ```bash
   ollama serve
   ```

## Build & Install

```bash
# Clone and navigate
cd /home/pefman/git/sidekick

# Build
make build
# or
./build.sh
# or
go build -o sidekick

# Install system-wide (optional, requires sudo)
sudo ./sidekick install
```

## Usage Examples

### 1. Scan current directory
```bash
./sidekick scan
```

### 2. Scan specific path
```bash
./sidekick scan /path/to/project
```

### 3. Use different model
```bash
./sidekick scan --model deepseek-coder
```

### 4. Verbose mode
```bash
./sidekick scan --verbose
```

### 5. Test with examples
```bash
./sidekick scan examples/
# or
make demo
```

## After Installation

Once installed with `sudo ./sidekick install`, you can use it from anywhere:

```bash
sidekick scan ~/my-project
sidekick scan --model codellama --verbose
```

## Troubleshooting

### "Failed to connect to Ollama"
Make sure Ollama is running:
```bash
ollama serve
```

### "Model not found"
Pull the model first:
```bash
ollama pull codellama
```

### Permission denied during install
Use sudo:
```bash
sudo ./sidekick install
```

## Recommended Models

- **codellama**: Best for general code analysis (7B-34B parameters)
- **deepseek-coder**: Great for security-focused analysis (6.7B-33B parameters)
- **llama2**: General purpose, good fallback

Larger models provide better analysis but require more resources.
