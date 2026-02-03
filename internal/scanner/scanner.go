package scanner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Confidence     string `json:"confidence,omitempty"`    // HIGH, MEDIUM, LOW
	IssueID        string `json:"issue_id,omitempty"`      // e.g., "CWE-89", "OWASP-A03"
	SuggestedFix   string `json:"suggested_fix,omitempty"` // Code to replace vulnerable code
	FixAvailable   bool   `json:"fix_available,omitempty"` // Whether LLM provided a fix
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
	if s.scanType == "triad" {
		result, err := s.scanTriadFiles(files)
		if err != nil {
			return nil, err
		}
		return []ScanResult{result}, nil
	}

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

		// Fix JSON string escaping issues (newlines in string values)
		findings = fixJSONStringEscaping(findings)

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

type triadStaticFinding struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
}

type triadVulnerability struct {
	Type           string `json:"type"`
	File           string `json:"file"`
	Line           int    `json:"line"`
	Evidence       string `json:"evidence"`
	Recommendation string `json:"recommendation"`
}

type triadReport struct {
	FinalSeverity   string               `json:"final_severity"`
	Confidence      string               `json:"confidence"`
	Vulnerabilities []triadVulnerability `json:"vulnerabilities"`
	Summary         string               `json:"summary,omitempty"`
}

func (s *Scanner) scanTriadFiles(files []string) (ScanResult, error) {
	result := ScanResult{
		FilePath:    "triad:multi",
		RawFindings: "",
		HasIssues:   false,
		Issues:      []SecurityIssue{},
	}

	if len(files) == 0 {
		return result, nil
	}

	codeByFile := make(map[string]string)
	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return result, fmt.Errorf("failed to read file: %w", err)
		}
		if len(content) == 0 || len(content) > 100000 {
			continue
		}
		codeByFile[filePath] = string(content)
	}

	if len(codeByFile) == 0 {
		return result, nil
	}

	staticFindings := runTriadStaticAnalysis(codeByFile)
	sharedContext := s.buildTriadSharedContext(codeByFile, staticFindings)

	var lastReport triadReport
	var summary string

	for round := 1; round <= 3; round++ {
		attackerPrompt := s.getTriadAttackerPrompt(sharedContext, summary, round)
		attackerResp, err := s.client.Generate(s.modelName, attackerPrompt)
		if err != nil {
			return result, fmt.Errorf("attacker pass failed: %w", err)
		}
		s.logDebug(fmt.Sprintf("TRIAD ROUND %d: ATTACKER PROMPT", round), attackerPrompt)
		s.logDebug(fmt.Sprintf("TRIAD ROUND %d: ATTACKER RESPONSE", round), attackerResp)

		defenderPrompt := s.getTriadDefenderPrompt(sharedContext, summary, attackerResp, round)
		defenderResp, err := s.client.Generate(s.modelName, defenderPrompt)
		if err != nil {
			return result, fmt.Errorf("defender pass failed: %w", err)
		}
		s.logDebug(fmt.Sprintf("TRIAD ROUND %d: DEFENDER PROMPT", round), defenderPrompt)
		s.logDebug(fmt.Sprintf("TRIAD ROUND %d: DEFENDER RESPONSE", round), defenderResp)

		auditorPrompt := s.getTriadAuditorPrompt(sharedContext, summary, attackerResp, defenderResp, round)
		auditorResp, err := s.client.Generate(s.modelName, auditorPrompt)
		if err != nil {
			return result, fmt.Errorf("auditor pass failed: %w", err)
		}
		s.logDebug(fmt.Sprintf("TRIAD ROUND %d: AUDITOR PROMPT", round), auditorPrompt)
		s.logDebug(fmt.Sprintf("TRIAD ROUND %d: AUDITOR RESPONSE", round), auditorResp)

		auditorResp = stripMarkdownCodeFences(auditorResp)
		auditorResp = fixJSONStringEscaping(auditorResp)

		if err := json.Unmarshal([]byte(auditorResp), &lastReport); err != nil {
			return result, fmt.Errorf("auditor response parse failed: %w. Raw output: %s", err, auditorResp)
		}

		summary = strings.TrimSpace(lastReport.Summary)
		if summary == "" {
			summary = truncateText(auditorResp, 1200)
		}

		if strings.EqualFold(lastReport.Confidence, "HIGH") {
			break
		}
	}

	finalJSON, err := json.MarshalIndent(lastReport, "", "  ")
	if err != nil {
		return result, fmt.Errorf("failed to serialize final report: %w", err)
	}

	result.RawFindings = string(finalJSON)
	result.HasIssues = len(lastReport.Vulnerabilities) > 0
	return result, nil
}

func runTriadStaticAnalysis(codeByFile map[string]string) []triadStaticFinding {
	var findings []triadStaticFinding

	for filePath, content := range codeByFile {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lineNum := i + 1
			trimmed := strings.TrimSpace(line)

			if strings.Contains(trimmed, "exec.Command") {
				findings = append(findings, triadStaticFinding{File: filePath, Line: lineNum, Pattern: "exec.Command", Severity: "High"})
			}
			if strings.Contains(trimmed, "tls.Config") && strings.Contains(trimmed, "InsecureSkipVerify") && strings.Contains(trimmed, "true") {
				findings = append(findings, triadStaticFinding{File: filePath, Line: lineNum, Pattern: "tls.Config{InsecureSkipVerify: true}", Severity: "High"})
			}
			if looksLikeSQLConcat(trimmed) {
				findings = append(findings, triadStaticFinding{File: filePath, Line: lineNum, Pattern: "SQL string concatenation", Severity: "High"})
			}
			if strings.Contains(trimmed, "http.ListenAndServe") {
				findings = append(findings, triadStaticFinding{File: filePath, Line: lineNum, Pattern: "http.ListenAndServe", Severity: "Medium"})
			}
			if looksLikeFmtSprintfUserInput(trimmed) {
				findings = append(findings, triadStaticFinding{File: filePath, Line: lineNum, Pattern: "fmt.Sprintf with user input", Severity: "Medium"})
			}
		}
	}

	return findings
}

func looksLikeSQLConcat(line string) bool {
	upper := strings.ToUpper(line)
	if !(strings.Contains(upper, "SELECT") || strings.Contains(upper, "INSERT") || strings.Contains(upper, "UPDATE") || strings.Contains(upper, "DELETE")) {
		return false
	}
	return strings.Contains(line, "+") || strings.Contains(line, "fmt.Sprintf")
}

func looksLikeFmtSprintfUserInput(line string) bool {
	if !strings.Contains(line, "fmt.Sprintf") {
		return false
	}
	userHints := []string{"r.", "req", "request", "user", "input", "param", "query", "form"}
	for _, hint := range userHints {
		if strings.Contains(strings.ToLower(line), hint) {
			return true
		}
	}
	return false
}

func (s *Scanner) buildTriadSharedContext(codeByFile map[string]string, findings []triadStaticFinding) string {
	var codeBuilder strings.Builder
	for filePath, content := range codeByFile {
		codeBuilder.WriteString(fmt.Sprintf("FILE: %s\n", filePath))
		codeBuilder.WriteString(addLineNumbers(content))
		codeBuilder.WriteString("\n")
	}

	findingsJSON, _ := json.MarshalIndent(findings, "", "  ")
	assumptions := "Assume code runs as a network service. User input may be untrusted. External dependencies may be attacker-controlled."

	context := fmt.Sprintf("CODE:\n%s\nSTATIC_FINDINGS:\n%s\nASSUMPTIONS:\n%s\n", codeBuilder.String(), string(findingsJSON), assumptions)
	return truncateText(context, 16000)
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen]
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

// fixJSONStringEscaping fixes unescaped newlines in JSON string values
func fixJSONStringEscaping(jsonStr string) string {
	// Simple approach: Replace literal newlines within quoted strings with \n
	// This handles cases where LLM outputs multi-line string values
	var result strings.Builder
	inString := false
	escaped := false

	for i := 0; i < len(jsonStr); i++ {
		ch := jsonStr[i]

		if escaped {
			result.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			result.WriteByte(ch)
			continue
		}

		if ch == '"' {
			inString = !inString
			result.WriteByte(ch)
			continue
		}

		// Replace actual newlines in strings with \n
		if inString && ch == '\n' {
			result.WriteString("\\n")
			continue
		}

		// Replace tabs in strings with \t
		if inString && ch == '\t' {
			result.WriteString("\\t")
			continue
		}

		result.WriteByte(ch)
	}

	return result.String()
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

// ReviewFindings implements interactive review mode for security findings
func ReviewFindings(findings []SecurityIssue, filePath string, client *ollama.Client, modelName string) error {
	if len(findings) == 0 {
		fmt.Println("No findings to review.")
		return nil
	}

	// Sort findings by line_start in DESCENDING order
	// This way we apply fixes from bottom to top, preventing line number shifts
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].LineStart > findings[j].LineStart
	})

	reader := bufio.NewReader(os.Stdin)
	currentIdx := 0
	appliedFixes := make(map[int]bool)
	backupCreated := false

	// Read file content once
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	lines := strings.Split(string(content), "\n")

	for {
		issue := findings[currentIdx]

		// Clear screen
		fmt.Print("\033[H\033[2J")

		// Header
		fmt.Printf("\n\033[38;5;208m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")
		fmt.Printf("\033[38;5;208m📋 Review Mode\033[0m - Finding %d of %d\n", currentIdx+1, len(findings))
		fmt.Printf("\033[38;5;208m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n\n")

		// Display issue details
		severityColor := getSeverityColor(issue.Severity)
		fmt.Printf("%s %s: %s\033[0m\n", severityColor, issue.Severity, issue.Title)
		fmt.Printf("📁 File: \033[36m%s\033[0m\n", filepath.Base(filePath))
		fmt.Printf("📍 Lines: \033[36m%d-%d\033[0m", issue.LineStart, issue.LineEnd)
		if issue.Confidence != "" {
			fmt.Printf(" | Confidence: %s", issue.Confidence)
		}
		if issue.IssueID != "" {
			fmt.Printf(" | %s", issue.IssueID)
		}
		fmt.Println("\n")

		fmt.Printf("📝 Description:\n%s\n\n", wrapText(issue.Description, 70))
		fmt.Printf("💡 Recommendation:\n%s\n\n", wrapText(issue.Recommendation, 70))

		// Show code context
		fmt.Printf("\033[38;5;208m━━━ Current Code ━━━\033[0m\n")
		showCodeContext(lines, issue.LineStart, issue.LineEnd)

		// Show fix status and diff
		if appliedFixes[currentIdx] {
			fmt.Printf("\n\033[38;5;82m✓ Fix already applied to this issue\033[0m\n")
		} else if issue.FixAvailable {
			fmt.Printf("\n\033[38;5;82m✓ Suggested fix available\033[0m\n")

			// Show diff by default if reasonable size
			original := string(content)
			diffLines := len(strings.Split(original, "\n")) + len(strings.Split(issue.SuggestedFix, "\n"))
			if diffLines <= 100 {
				showDiff(original, issue.SuggestedFix)
			} else {
				fmt.Printf("\n\033[38;5;203m(Diff too large - use [s] to show)\033[0m\n")
			}
		} else {
			fmt.Printf("\n\033[38;5;203m⚠ No automatic fix available - manual review required\033[0m\n")
		}

		// Action menu
		fmt.Printf("\n\033[38;5;208m━━━ Actions ━━━\033[0m\n")
		if issue.FixAvailable && !appliedFixes[currentIdx] {
			fmt.Println("  [a] Apply fix")
			fmt.Println("  [s] Show diff")
		}
		fmt.Println("  [i] Ignore (skip this finding)")
		if currentIdx < len(findings)-1 {
			fmt.Println("  [n] Next finding")
		}
		if currentIdx > 0 {
			fmt.Println("  [p] Previous finding")
		}
		fmt.Println("  [q] Quit review mode")
		fmt.Printf("\n\033[38;5;208mChoice:\033[0m ")

		// Read input
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		choice := strings.ToLower(strings.TrimSpace(input))

		switch choice {
		case "a":
			if !issue.FixAvailable {
				fmt.Println("\n\033[38;5;203m⚠ No fix available for this issue\033[0m")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
				continue
			}
			if appliedFixes[currentIdx] {
				fmt.Println("\n\033[38;5;203m⚠ Fix already applied\033[0m")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
				continue
			}

			// Create backup on first fix
			if !backupCreated {
				backupPath := filePath + ".backup"
				if err := os.WriteFile(backupPath, content, 0644); err != nil {
					fmt.Printf("\n\033[38;5;203m✗ Failed to create backup: %v\033[0m\n", err)
					fmt.Print("Press Enter to continue...")
					reader.ReadString('\n')
					continue
				}
				backupCreated = true
				fmt.Printf("\n\033[38;5;82m✓ Backup created: %s\033[0m\n", backupPath)
			}

			// Use the suggested fix directly (no validation)
			issue.SuggestedFix = extractCodeFromResponse(issue.SuggestedFix)

			// Count lines in the fix and adjust line_end if needed
			fixLineCount := len(strings.Split(strings.TrimSpace(issue.SuggestedFix), "\n"))
			originalLineCount := issue.LineEnd - issue.LineStart + 1

			// If fix has more lines than original, expand the range to match
			// This handles cases where LLM initially identified single line but fix spans multiple
			if fixLineCount > originalLineCount {
				issue.LineEnd = issue.LineStart + fixLineCount - 1
				// Make sure we don't go past end of file
				if issue.LineEnd > len(lines) {
					issue.LineEnd = len(lines)
				}
			}

			// Apply the fix to the file
			if err := applyFix(filePath, issue); err != nil {
				fmt.Printf("\n\033[38;5;203m✗ Failed to apply fix: %v\033[0m\n", err)
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
				continue
			}

			// Mark as applied and reload content for next fixes
			appliedFixes[currentIdx] = true
			content, err = os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("\n\033[38;5;203m✗ Failed to reload file: %v\033[0m\n", err)
				return err
			}
			lines = strings.Split(string(content), "\n")

			fmt.Printf("\n\033[38;5;82m✓ Fix applied successfully!\033[0m\n")

			// Auto-advance to next finding
			if currentIdx < len(findings)-1 {
				currentIdx++
				fmt.Println("\033[38;5;82mMoving to next finding...\033[0m")
				time.Sleep(800 * time.Millisecond)
			} else {
				fmt.Println("\n\033[38;5;82mAll findings reviewed!\033[0m")
				return nil
			}

		case "s":
			if !issue.FixAvailable {
				fmt.Println("\n\033[38;5;203m⚠ No fix available to show\033[0m")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
				continue
			}

			// Extract original code
			original := extractLines(lines, issue.LineStart, issue.LineEnd)
			showDiff(original, issue.SuggestedFix)
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "i":
			fmt.Printf("\n\033[38;5;82m✓ Ignoring this finding\033[0m\n")
			if currentIdx < len(findings)-1 {
				currentIdx++
			} else {
				fmt.Println("No more findings. Exiting review mode.")
				return nil
			}

		case "n":
			if currentIdx < len(findings)-1 {
				currentIdx++
			} else {
				fmt.Println("\nAlready at last finding")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
			}

		case "p":
			if currentIdx > 0 {
				currentIdx--
			} else {
				fmt.Println("\nAlready at first finding")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
			}

		case "q":
			fmt.Println("\n\033[38;5;208m👋 Exiting review mode\033[0m")
			return nil

		default:
			fmt.Println("\n\033[38;5;203m⚠ Invalid choice\033[0m")
			fmt.Print("Press Enter to continue...")
			reader.ReadString('\n')
		}
	}
}

// applyFix applies the suggested fix to the file
func applyFix(filePath string, issue SecurityIssue) error {
	if !issue.FixAvailable || issue.SuggestedFix == "" {
		return fmt.Errorf("no fix available")
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate and clamp line numbers (LLM sometimes gives inaccurate line numbers)
	if issue.LineStart < 1 {
		issue.LineStart = 1
	}
	if issue.LineStart > len(lines) {
		issue.LineStart = len(lines)
	}
	if issue.LineEnd < issue.LineStart {
		issue.LineEnd = issue.LineStart
	}
	if issue.LineEnd > len(lines) {
		issue.LineEnd = len(lines)
	}

	// Detect original indentation from the first line of the issue
	originalIndent := ""
	if issue.LineStart-1 < len(lines) && len(lines[issue.LineStart-1]) > 0 {
		line := lines[issue.LineStart-1]
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				originalIndent += string(ch)
			} else {
				// Stop at first non-whitespace character
				break
			}
		}
	}

	// Replace lines
	fixLines := strings.Split(strings.TrimSuffix(issue.SuggestedFix, "\n"), "\n")

	// Find minimum indentation in the fix (to detect relative indentation)
	minFixIndent := -1
	for _, fixLine := range fixLines {
		if strings.TrimSpace(fixLine) == "" {
			continue // Skip empty lines
		}
		indent := 0
		for _, ch := range fixLine {
			if ch == ' ' {
				indent++
			} else if ch == '\t' {
				indent += 4 // Treat tab as 4 spaces
			} else {
				break
			}
		}
		if minFixIndent == -1 || indent < minFixIndent {
			minFixIndent = indent
		}
	}
	if minFixIndent == -1 {
		minFixIndent = 0
	}

	// Apply original indentation while preserving relative indentation
	for i, fixLine := range fixLines {
		trimmed := strings.TrimSpace(fixLine)
		if trimmed == "" {
			fixLines[i] = ""
			continue
		}

		// Calculate current indentation
		currentIndent := 0
		for _, ch := range fixLine {
			if ch == ' ' {
				currentIndent++
			} else if ch == '\t' {
				currentIndent += 4
			} else {
				break
			}
		}

		// Calculate relative indentation from minimum
		relativeIndent := currentIndent - minFixIndent

		// Build new line with original base indent + relative indent + content
		var newIndent string
		if strings.Contains(originalIndent, "\t") {
			// Use tabs if original uses tabs
			tabCount := relativeIndent / 4
			spaceCount := relativeIndent % 4
			newIndent = originalIndent + strings.Repeat("\t", tabCount) + strings.Repeat(" ", spaceCount)
		} else {
			// Use spaces
			newIndent = originalIndent + strings.Repeat(" ", relativeIndent)
		}

		fixLines[i] = newIndent + trimmed
	}

	// Build new content
	var newLines []string
	newLines = append(newLines, lines[:issue.LineStart-1]...) // Lines before issue
	newLines = append(newLines, fixLines...)                  // Fixed code
	if issue.LineEnd < len(lines) {
		newLines = append(newLines, lines[issue.LineEnd:]...) // Lines after issue
	}

	newContent := strings.Join(newLines, "\n")

	// Write fixed content
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// showDiff displays before/after with color coding
func showDiff(original, fixed string) {
	fmt.Printf("\n\033[38;5;208m━━━ Diff Preview ━━━\033[0m\n\n")

	fmt.Printf("\033[38;5;203m--- Original ---\033[0m\n")
	originalLines := strings.Split(strings.TrimSpace(original), "\n")
	for _, line := range originalLines {
		fmt.Printf("\033[38;5;203m- %s\033[0m\n", line)
	}

	fmt.Printf("\n\033[38;5;82m+++ Fixed ---\033[0m\n")
	fixedLines := strings.Split(strings.TrimSpace(fixed), "\n")
	for _, line := range fixedLines {
		fmt.Printf("\033[38;5;82m+ %s\033[0m\n", line)
	}
}

// Helper functions

func getSeverityColor(severity string) string {
	switch severity {
	case "CRITICAL":
		return "\033[38;5;196m🔴" // Bright red
	case "HIGH":
		return "\033[38;5;208m🟠" // Orange
	case "MEDIUM":
		return "\033[38;5;226m🟡" // Yellow
	case "LOW":
		return "\033[38;5;82m🟢" // Green
	default:
		return "\033[0m"
	}
}

func showCodeContext(lines []string, lineStart, lineEnd int) {
	contextBefore := 2
	contextAfter := 2

	start := maxInt(0, lineStart-1-contextBefore)
	end := minInt(len(lines), lineEnd+contextAfter)

	for i := start; i < end; i++ {
		lineNum := i + 1
		line := lines[i]

		if lineNum >= lineStart && lineNum <= lineEnd {
			// Highlight vulnerable lines
			fmt.Printf("\033[38;5;203m%4d | %s\033[0m\n", lineNum, line)
		} else {
			// Context lines
			fmt.Printf("\033[38;5;240m%4d | %s\033[0m\n", lineNum, line)
		}
	}
}

func extractLines(lines []string, lineStart, lineEnd int) string {
	if lineStart < 1 || lineStart > len(lines) {
		return ""
	}
	if lineEnd < lineStart || lineEnd > len(lines) {
		lineEnd = len(lines)
	}

	extracted := lines[lineStart-1 : lineEnd]
	return strings.Join(extracted, "\n")
}

func wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var result strings.Builder
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}

		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += wordLen

		if i < len(words)-1 && lineLen > 0 {
			// Continue
		}
	}

	return result.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractCodeFromResponse parses the LLM's structured response to get just the code part
func extractCodeFromResponse(response string) string {
	// Look for CODE: section
	lines := strings.Split(response, "\n")
	inCodeSection := false
	var codeLines []string

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "CODE:") {
			inCodeSection = true
			continue
		}
		if inCodeSection {
			// Skip markdown code fences
			trimmed := strings.TrimSpace(line)
			if trimmed == "```" || strings.HasPrefix(trimmed, "```go") || strings.HasPrefix(trimmed, "```python") || strings.HasPrefix(trimmed, "```java") || strings.HasPrefix(trimmed, "```javascript") || strings.HasPrefix(trimmed, "```") {
				continue
			}
			codeLines = append(codeLines, line)
		}
	}

	// If no CODE: marker found, strip markdown fences from entire response
	if len(codeLines) == 0 {
		result := strings.TrimSpace(response)
		// Remove markdown code fences
		result = strings.ReplaceAll(result, "```go", "")
		result = strings.ReplaceAll(result, "```python", "")
		result = strings.ReplaceAll(result, "```java", "")
		result = strings.ReplaceAll(result, "```javascript", "")
		result = strings.ReplaceAll(result, "```", "")
		return strings.TrimSpace(result)
	}

	return strings.TrimSpace(strings.Join(codeLines, "\n"))
}
