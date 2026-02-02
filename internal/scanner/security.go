package scanner

import (
	"fmt"
	"path/filepath"
)

// Stage 1: Context Analysis
func (s *Scanner) analyzeContext(filename, content string) (string, error) {
	return s.client.Generate(s.modelName, s.getContextPrompt(filename, content))
}

func (s *Scanner) getContextPrompt(filename, content string) string {
	return fmt.Sprintf(`Analyze the context of this code file to help guide a security scan.

FILE: %s
CODE (with line numbers):
%s

CRITICAL: Output ONLY valid JSON, no other text.

Output format:
{
  "language": "Programming language name",
  "version": "Version if evident, otherwise 'unknown'",
  "frameworks": ["List", "of", "frameworks"],
  "libraries": ["List", "of", "libraries"],
  "purpose": "Brief description of code purpose",
  "data_handling": ["user_input", "database", "filesystem", "network"],
  "security_concerns": ["Key security risks for this tech stack"]
}

Note: The code has line numbers prefixed (e.g., "1 | package main"). These are the actual line numbers - use them for precise vulnerability reporting.`, filename, content)
}

// Stage 2: Security Scan with Context
func (s *Scanner) scanWithContext(filename, content, context string) (string, error) {
	return s.client.Generate(s.modelName, s.getScanPrompt(filename, content, context))
}

func (s *Scanner) getScanPrompt(filename, content, context string) string {
	_ = filepath.Base("") // keep import
	return fmt.Sprintf(`Based on this context analysis:

%s

Now perform a thorough security scan of the code:

FILE: %s
CODE (with line numbers):
%s

IMPORTANT: The code has line numbers prefixed (e.g., "42 | if err != nil"). Use these EXACT line numbers in your response.

CRITICAL INSTRUCTIONS:
- Output ONLY raw JSON
- NO markdown code fences (no triple-backtick json markers)
- NO explanatory text before or after the JSON
- Start your response directly with { and end with }

Output format (JSON only):
{
  "findings": [
    {
      "severity": "CRITICAL|HIGH|MEDIUM|LOW",
      "title": "Brief title (e.g., 'SQL Injection', 'Hardcoded Credentials')",
      "description": "Detailed explanation of the vulnerability",
      "line_start": <number>,
      "line_end": <number>,
      "recommendation": "How to fix this issue",
      "confidence": "HIGH|MEDIUM|LOW",
      "issue_id": "CWE-XXX or OWASP-AXX (optional)"
    }
  ]
}

Rules:
- severity: CRITICAL, HIGH, MEDIUM, or LOW
- line_start and line_end: use the EXACT numbers from the prefixed code
- confidence: HIGH (certain), MEDIUM (likely), LOW (possible)
- issue_id: CWE/OWASP identifier if applicable (can be omitted)
- If no vulnerabilities found, output: {"findings": []}
- Your response must be valid JSON that can be parsed directly`, context, filename, content)
}
