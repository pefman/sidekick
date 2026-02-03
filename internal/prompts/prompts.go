package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed custom/*.txt
var promptFS embed.FS

type CustomPromptData struct {
	Mode       string
	UserPrompt string
	FilePath   string
	Code       string
}

func RenderCustomPrompt(data CustomPromptData) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(data.Mode))
	if mode == "" {
		mode = "ask"
	}
	if mode != "ask" && mode != "edit" && mode != "plan" {
		mode = "ask"
	}

	path := fmt.Sprintf("custom/%s.txt", mode)
	tmplBytes, err := promptFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read prompt template: %w", err)
	}

	tmpl, err := template.New(path).Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render prompt template: %w", err)
	}

	return buf.String(), nil
}
