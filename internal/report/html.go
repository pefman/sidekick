package report

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/pefman/sidekick/internal/scanner"
)

type HTMLReport struct {
	Timestamp       string
	ScanPath        string
	Model           string
	TotalFiles      int
	FilesWithIssues int
	Results         []scanner.ScanResult
	GenerationTime  string
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sidekick Report</title>
    <style>
        body { font-family: Arial, sans-serif; background: #0a0a0a; color: #e0e0e0; margin: 0; padding: 24px; }
        .container { max-width: 1100px; margin: 0 auto; background: #111; border: 1px solid #222; border-radius: 6px; }
        .header { padding: 20px 24px; border-bottom: 1px solid #222; color: #ff7e00; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; padding: 20px 24px; }
        .card { background: #151515; padding: 16px; border: 1px solid #222; }
        .content { padding: 20px 24px; }
        .file { margin-bottom: 16px; border: 1px solid #222; }
        .file-header { background: #1a1a1a; color: #ff7e00; padding: 10px 12px; }
        .findings { padding: 12px; }
        .footer { padding: 16px; text-align: center; color: #777; border-top: 1px solid #222; }
        pre { white-space: pre-wrap; }
    </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>Sidekick Report</h2>
    </div>
    <div class="summary">
      <div class="card">Files Scanned: {{.TotalFiles}}</div>
      <div class="card">Files With Findings: {{.FilesWithIssues}}</div>
      <div class="card">Model: {{.Model}}</div>
    </div>
    <div class="content">
      {{range .Results}}
      {{if .HasIssues}}
      <div class="file">
        <div class="file-header">{{.FilePath}}</div>
        <div class="findings"><pre>{{.RawFindings}}</pre></div>
      </div>
      {{end}}
      {{end}}
    </div>
    <div class="footer">Generated {{.GenerationTime}}</div>
  </div>
</body>
</html>`

func GenerateHTML(results []scanner.ScanResult, scanPath, model string, totalFiles int, outputPath string) error {
	filesWithIssues := 0
	for _, result := range results {
		if result.HasIssues {
			filesWithIssues++
		}
	}

	report := HTMLReport{
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
		ScanPath:        scanPath,
		Model:           model,
		TotalFiles:      totalFiles,
		FilesWithIssues: filesWithIssues,
		Results:         results,
		GenerationTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, report); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func GetDefaultReportPath(scanPath string) string {
	timestamp := time.Now().Format("20060102-150405")
	baseName := filepath.Base(scanPath)
	if baseName == "." || baseName == "/" {
		baseName = "scan"
	}
	return fmt.Sprintf("sidekick-report-%s-%s.html", baseName, timestamp)
}
