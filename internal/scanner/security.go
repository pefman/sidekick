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
      "issue_id": "CWE-XXX or OWASP-AXX (optional)",
      "fix_available": true|false,
      "suggested_fix": "Complete replacement code for lines line_start to line_end (only if fix_available is true)"
    }
  ]
}

Rules:
- severity: CRITICAL, HIGH, MEDIUM, or LOW
- line_start and line_end: use the EXACT numbers from the prefixed code
- confidence: HIGH (certain), MEDIUM (likely), LOW (possible)
- issue_id: CWE/OWASP identifier if applicable (can be omitted)
- fix_available: true if you can provide a code fix, false if it requires manual intervention (e.g., architecture changes, hardcoded secrets that need external config)
- suggested_fix: ONLY if fix_available is true, provide the complete replacement code for the vulnerable lines
- If no vulnerabilities found, output: {"findings": []}
- Your response must be valid JSON that can be parsed directly`, context, filename, content)
}

func (s *Scanner) getTriadAttackerPrompt(sharedContext, summary string, round int) string {
	return fmt.Sprintf(`You are the ATTACKER in round %d.

Shared context:
%s

Prior summary (if any):
%s

Task:
- Assume a hostile environment.
- Identify concrete exploit scenarios based on the code and static findings.
- Include preconditions, exploitation steps, and impact.

Output format:
- Bullet list.
- Reference file names and line numbers where possible.
`, round, sharedContext, summary)
}

func (s *Scanner) getTriadDefenderPrompt(sharedContext, summary, attackerResponse string, round int) string {
	return fmt.Sprintf(`You are the DEFENDER in round %d.

Shared context:
%s

Prior summary (if any):
%s

Attacker claims:
%s

Task:
- Challenge attacker claims with evidence.
- Identify false positives or mitigating factors.
- Note Go runtime protections or deployment assumptions.

Output format:
- Bullet list of rebuttals and mitigations.
`, round, sharedContext, summary, attackerResponse)
}

func (s *Scanner) getTriadAuditorPrompt(sharedContext, summary, attackerResponse, defenderResponse string, round int) string {
	return fmt.Sprintf(`You are the AUDITOR in round %d.

Shared context:
%s

Prior summary (if any):
%s

Attacker claims:
%s

Defender rebuttals:
%s

Task:
- Resolve disagreements using evidence from the code and findings.
- Provide final severity and confidence.
- Prefer evidence over speculation.

CRITICAL INSTRUCTIONS:
- Output ONLY raw JSON.
- No markdown fences.
- Start with { and end with }.

Output JSON format:
{
  "final_severity": "Low|Medium|High|Critical",
  "confidence": "Low|Medium|High",
  "summary": "Short summary for next round",
  "vulnerabilities": [
    {
      "type": "Short vulnerability name",
      "file": "path/to/file.go",
      "line": 123,
      "evidence": "Concrete evidence from code",
      "recommendation": "Specific fix recommendation"
    }
  ]
}
`, round, sharedContext, summary, attackerResponse, defenderResponse)
}
