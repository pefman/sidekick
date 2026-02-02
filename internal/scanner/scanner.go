package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pefman/sidekick/internal/ollama"
	"github.com/pefman/sidekick/internal/ui"
)

type Scanner struct {
	client       *ollama.Client
	modelName    string
	debug        bool
	debugFile    *os.File
	scanType     string
	customPrompt string
}

type ScanResult struct {
	FilePath    string
	RawFindings string // Only used for custom prompts (unstructured)
	HasIssues   bool
	Issues      []SecurityIssue // Primary data structure for security scans
}

type SecurityIssue struct {
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	LineStart      int    `json:"line_start"`
	LineEnd        int    `json:"line_end"`
	Recommendation string `json:"recommendation"`
	Confidence     string `json:"confidence,omitempty"` // HIGH, MEDIUM, LOW
	IssueID        string `json:"issue_id,omitempty"`   // e.g., "CWE-89", "OWASP-A03"
}

func NewScanner(client *ollama.Client, modelName string, debug bool, scanType, customPrompt string) *Scanner {
	var debugFile *os.File
	if debug {
		// Create debug file with timestamp
		timestamp := time.Now().Format("20060102-150405")
		debugPath := fmt.Sprintf("sidekick-debug-%s.log", timestamp)
		f, err := os.Create(debugPath)
		if err == nil {
			debugFile = f
			fmt.Printf("\n🔍 Debug logging enabled: %s\n\n", debugPath)
		}
	}

	return &Scanner{
		client:       client,
		modelName:    modelName,
		debug:        debug,
		debugFile:    debugFile,
		scanType:     scanType,
		customPrompt: customPrompt,
	}
}

func (s *Scanner) logDebug(title, content string) {
	if s.debug && s.debugFile != nil {
		fmt.Fprintf(s.debugFile, "\n%s\n%s\n%s\n%s\n",
			strings.Repeat("=", 80),
			title,
			strings.Repeat("=", 80),
			content)
	}
}

func (s *Scanner) ScanFiles(files []string) ([]ScanResult, error) {
	results := make([]ScanResult, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Worker pool - limit concurrent scans to 3
	workers := 3
	jobs := make(chan int, len(files))

	// Progress tracking with single spinner
	var completed int
	var progressMu sync.Mutex
	spinner := ui.NewSpinner("")

	// Helper to update spinner safely
	updateSpinner := func(msg string) {
		progressMu.Lock()
		spinner.UpdateMessage(msg)
		progressMu.Unlock()
	}

	// Start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				file := files[i]

				// Update progress
				progressMu.Lock()
				completed++
				current := completed
				if current == 1 {
					spinner.Start()
				}
				progressMu.Unlock()

				// Calculate total stages (security = 3 stages: read, context, scan; custom = 2 stages: read, analysis)
				stagesPerFile := 2
				if s.scanType == "security" {
					stagesPerFile = 3
				}
				totalStages := len(files) * stagesPerFile
				startStage := (current - 1) * stagesPerFile

				// Pass spinner update function to scanFile
				result, err := s.scanFileWithProgress(file, startStage, totalStages, stagesPerFile, updateSpinner)

				if err != nil {
					progressMu.Lock()
					spinner.Stop()
					fmt.Fprintf(os.Stderr, "⚠️  Failed to scan %s: %v\n", file, err)
					spinner.Start()
					progressMu.Unlock()
					continue
				}

				// Always append results (even with no issues)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}()
	}

	// Queue all files
	for i := range files {
		jobs <- i
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	spinner.Stop()

	return results, nil
}

func (s *Scanner) Close() {
	if s.debugFile != nil {
		s.debugFile.Close()
	}
}

func (s *Scanner) scanFileWithProgress(filePath string, startStage, totalStages, stagesPerFile int, updateStatus func(string)) (ScanResult, error) {
	result := ScanResult{
		FilePath: filePath,
		Issues:   make([]SecurityIssue, 0),
	}

	fileName := filepath.Base(filePath)
	currentStage := startStage + 1

	// Reading file
	updateStatus(fmt.Sprintf("[%d/%d] Reading %s", currentStage, totalStages, fileName))
	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to read file: %w", err)
	}

	// Skip empty or very large files
	if len(content) == 0 {
		return result, nil
	}
	if len(content) > 100000 { // Skip files larger than 100KB
		return result, nil
	}

	if s.scanType == "security" {
		// Stage 1: Context Analysis
		currentStage++
		updateStatus(fmt.Sprintf("[%d/%d] Identifying language/frameworks in %s", currentStage, totalStages, fileName))
		// Add line numbers to code for precise references
		numberedContent := addLineNumbers(string(content))
		contextAnalysis, err := s.analyzeContext(filePath, numberedContent)
		if err != nil {
			return result, fmt.Errorf("context analysis failed: %w", err)
		}

		s.logDebug("STAGE 1: CONTEXT ANALYSIS PROMPT", s.getContextPrompt(filePath, numberedContent))
		s.logDebug("STAGE 1: CONTEXT ANALYSIS RESPONSE", contextAnalysis)

		// Strip markdown if present and validate JSON (optional - we pass raw to Stage 2)
		contextAnalysis = stripMarkdownCodeFences(contextAnalysis)

		// Stage 2: Targeted Scan
		currentStage++
		updateStatus(fmt.Sprintf("[%d/%d] Checking for vulnerabilities in %s", currentStage, totalStages, fileName))
		// Use numbered content so LLM can reference exact lines
		findings, err := s.scanWithContext(filePath, numberedContent, contextAnalysis)
		if err != nil {
			return result, fmt.Errorf("security scan failed: %w", err)
		}

		s.logDebug("STAGE 2: SECURITY SCAN PROMPT", s.getScanPrompt(filePath, string(content), contextAnalysis))
		s.logDebug("STAGE 2: SECURITY SCAN RESPONSE", findings)

		// Strip markdown code fences if present
		findings = stripMarkdownCodeFences(findings)

		// Parse JSON response
		var jsonResponse struct {
			Findings []SecurityIssue `json:"findings"`
		}

		if err := json.Unmarshal([]byte(findings), &jsonResponse); err != nil {
			return result, fmt.Errorf("failed to parse JSON response: %w. Raw output: %s", err, findings)
		}

		result.Issues = jsonResponse.Findings
		result.HasIssues = len(jsonResponse.Findings) > 0

		// Render findings to text for display
		result.RawFindings = s.renderFindings(jsonResponse.Findings)

		return result, nil
	} else {
		// Custom prompt - simpler flow
		prompt := s.createCustomPrompt(filePath, string(content))

		s.logDebug("CUSTOM PROMPT", prompt)

		currentStage++
		updateStatus(fmt.Sprintf("[%d/%d] Running custom analysis on %s", currentStage, totalStages, fileName))
		response, err := s.client.Generate(s.modelName, prompt)
		if err != nil {
			return result, fmt.Errorf("analysis failed: %w", err)
		}

		s.logDebug("CUSTOM RESPONSE", response)

		result.RawFindings = response
		result.HasIssues = strings.TrimSpace(response) != ""
		result.Issues = []SecurityIssue{} // Keep empty for custom prompts
	}

	return result, nil
}

func (s *Scanner) scanFile(filePath string) (ScanResult, error) {
	// Legacy method - calls new method with no-op progress
	stagesPerFile := 2
	if s.scanType == "security" {
		stagesPerFile = 3
	}
	return s.scanFileWithProgress(filePath, 0, stagesPerFile, stagesPerFile, func(string) {})
}

// renderFindings converts structured SecurityIssue data to formatted text output
func (s *Scanner) renderFindings(issues []SecurityIssue) string {
	if len(issues) == 0 {
		return "No security issues found."
	}

	var output strings.Builder

	// Severity emoji mapping
	severityEmoji := map[string]string{
		"CRITICAL": "🔴",
		"HIGH":     "🟠",
		"MEDIUM":   "🟡",
		"LOW":      "🟢",
	}

	output.WriteString("===================================\n")
	output.WriteString("Security Analysis Report\n")
	output.WriteString("===================================\n\n")

	// Group by severity
	bySeverity := make(map[string][]SecurityIssue)
	for _, issue := range issues {
		bySeverity[issue.Severity] = append(bySeverity[issue.Severity], issue)
	}

	// Display in severity order
	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"} {
		if items, ok := bySeverity[sev]; ok && len(items) > 0 {
			for _, issue := range items {
				emoji := severityEmoji[sev]
				output.WriteString(fmt.Sprintf("%s %s: %s\n", emoji, sev, issue.Title))

				lineInfo := fmt.Sprintf("Line: %d", issue.LineStart)
				if issue.LineEnd != issue.LineStart {
					lineInfo = fmt.Sprintf("Lines: %d-%d", issue.LineStart, issue.LineEnd)
				}
				output.WriteString(fmt.Sprintf("   %s", lineInfo))

				// Add confidence if present
				if issue.Confidence != "" {
					output.WriteString(fmt.Sprintf(" | Confidence: %s", issue.Confidence))
				}
				// Add issue ID if present
				if issue.IssueID != "" {
					output.WriteString(fmt.Sprintf(" | %s", issue.IssueID))
				}
				output.WriteString("\n\n")

				output.WriteString(fmt.Sprintf("   Description:\n   %s\n\n", issue.Description))
				output.WriteString(fmt.Sprintf("   Recommendation:\n   %s\n\n", issue.Recommendation))
				output.WriteString("-----------------------------------\n\n")
			}
		}
	}

	// Summary
	output.WriteString("Summary:\n")
	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"} {
		if items, ok := bySeverity[sev]; ok && len(items) > 0 {
			emoji := severityEmoji[sev]
			output.WriteString(fmt.Sprintf("  %s %s: %d\n", emoji, sev, len(items)))
		}
	}

	return output.String()
}

// addLineNumbers prefixes each line with its line number
func addLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	var numbered strings.Builder
	for i, line := range lines {
		numbered.WriteString(fmt.Sprintf("%4d | %s\n", i+1, line))
	}
	return numbered.String()
}

// stripMarkdownCodeFences removes ```json and ``` wrappers if present
func stripMarkdownCodeFences(s string) string {
	s = strings.TrimSpace(s)
	// Remove ```json or ``` at start
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	// Remove ``` at end
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to count lines before a position
func countLines(content string, pos int) int {
	if pos >= len(content) {
		pos = len(content) - 1
	}
	return strings.Count(content[:pos], "\n") + 1
}
