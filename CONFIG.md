# Sidekick Configuration (Future Enhancement)

This file is a placeholder for future configuration options.

## Planned Features

```yaml
# .sidekick.yaml
ollama:
  url: "http://localhost:11434"
  default_model: "codellama"
  timeout: 300s

scan:
  max_file_size: 102400  # 100KB
  extensions:
    - .go
    - .js
    - .ts
    - .py
    - .java
    - .c
    - .cpp
    - .rs
    - .rb
    - .php
  
  exclude_dirs:
    - node_modules
    - vendor
    - .git
    - dist
    - build
  
  severity_threshold: "LOW"  # Only show issues at or above this level

output:
  format: "text"  # text, json, sarif
  color: true
  verbose: false

```

## Usage (Future)

Create a `.sidekick.yaml` in your project root or home directory:

```bash
# Project-specific config
touch .sidekick.yaml

# Global config
touch ~/.sidekick.yaml
```

Priority order:
1. Current directory `.sidekick.yaml`
2. Home directory `~/.sidekick.yaml`
3. Command-line flags
4. Default values
