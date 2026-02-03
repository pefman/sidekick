# Sidekick Configuration

Sidekick stores user settings at:

```
~/.sidekick/config.json
```

Example:

```json
{
  "default_model": "qwen2.5-coder:14b-instruct-q4",
  "ollama_url": "http://localhost:11434",
  "debug": false,
  "output_format": "text"
}
```

## Notes
- Use the **Settings** menu to update these values.
- CLI flags override config values for a single run.
