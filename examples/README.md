# Example Vulnerable Code

This directory contains example code with intentional security vulnerabilities for testing sidekick.

## Testing

To scan the example code:

```bash
# Scan the examples directory
./sidekick scan examples/

# Or scan with verbose output
./sidekick scan examples/ --verbose
```

## Vulnerabilities in vulnerable_code.go

The example file contains several common security issues:

1. **Hardcoded Credentials**: Database credentials stored in source code
2. **SQL Injection**: Unsanitized user input in SQL queries
3. **Command Injection**: User input passed directly to system commands
4. **Path Traversal**: No validation on file paths from user input

## Expected Results

Sidekick should detect and report these vulnerabilities with:
- Severity levels (CRITICAL, HIGH, MEDIUM, LOW)
- Line numbers where issues occur
- Recommendations for fixing each issue

## Note

**⚠️ WARNING**: This code contains intentional security vulnerabilities for demonstration purposes only. Never use this code in production!
