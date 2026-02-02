package scanner

import (
	"fmt"
)

func (s *Scanner) createCustomPrompt(filename, content string) string {
	prompt := fmt.Sprintf(`%s

Analyze this file: %s

%s

Provide your analysis in a clear and concise format.
`, s.customPrompt, filename, content)

	return prompt
}

// For custom prompts, we still parse as issues for now
// but could be enhanced to show raw output in the future
