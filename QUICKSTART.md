# Quick Start

## 1) Install Ollama
```bash
curl -fsSL https://ollama.ai/install.sh | sh
ollama serve
```

## 2) Pull a model
```bash
ollama pull qwen2.5-coder:14b-instruct-q4
```

## 3) Build and run
```bash
go build -o sidekick
./sidekick
```

## 4) Use the prompt UI
- Type your request on the first line
- Press **Tab** to switch **Ask / Edit / Plan**
- Use **↑/↓** to select Settings/Models/Help
- Press **Enter** to execute the prompt or open the selected menu

## 5) CLI scan
```bash
./sidekick scan /path/to/project
```
