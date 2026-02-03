package scanner

import (
	"fmt"
	"strings"

	"github.com/pefman/sidekick/internal/prompts"
)

func (s *Scanner) createCustomPrompt(filename, content string) string {
	mode, userPrompt := parseCustomPrompt(s.customPrompt)

	result, err := prompts.RenderCustomPrompt(prompts.CustomPromptData{
		Mode:       mode,
		UserPrompt: userPrompt,
		FilePath:   filename,
		Code:       content,
	})
	if err != nil {
		return fmt.Sprintf("%s\n\nFILE: %s\nCODE:\n%s\n", s.customPrompt, filename, content)
	}

	return result
}

func parseCustomPrompt(raw string) (string, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "ask", ""
	}

	lines := strings.SplitN(trimmed, "\n", 2)
	head := strings.TrimSpace(lines[0])
	if strings.HasPrefix(strings.ToUpper(head), "MODE:") {
		mode := strings.ToLower(strings.TrimSpace(head[len("MODE:"):]))
		body := ""
		if len(lines) > 1 {
			body = strings.TrimSpace(lines[1])
		}
		if mode == "ask" || mode == "edit" || mode == "plan" {
			return mode, body
		}
		return "ask", body
	}

	return "ask", trimmed
}

// For custom prompts, we still parse as issues for now
// but could be enhanced to show raw output in the future
