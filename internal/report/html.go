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
    <title>Sidekick Analysis Report</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Courier New', monospace;
            background: #0a0a0a;
            padding: 20px;
            color: #c0c0c0;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: #1a1a1a;
            border-radius: 4px;
            box-shadow: 0 0 30px rgba(255, 126, 0, 0.2);
            overflow: hidden;
            border: 1px solid #2a2a2a;
        }
        .header {
            background: #1a1a1a;
            border-bottom: 2px solid #ff7e00;
            color: #ff7e00;
            padding: 20px 40px;
        }
        .header h1 {
            font-size: 1em;
            margin-bottom: 3px;
            font-weight: normal;
            letter-spacing: 2px;
        }
        .header h1::before {
            content: '▸ ';
        }
        .header p {
            font-size: 0.85em;
            color: #666;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 15px;
            padding: 30px;
            background: #151515;
            border-bottom: 1px solid #2a2a2a;
        }
        .summary-card {
            background: #1a1a1a;
            padding: 20px;
            border: 1px solid #2a2a2a;
            text-align: center;
        }
        .summary-card h3 {
            font-size: 0.8em;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 10px;
            font-weight: normal;
        }
        .summary-card .number {
            font-size: 2.5em;
            font-weight: normal;
            margin: 10px 0;
        }
        .severity-critical { color: #ff3860; }
        .severity-high { color: #ff7e00; }
        .severity-medium { color: #ffdd57; }
        .severity-low { color: #33ccff; }
        .severity-total { color: #ff7e00; }
        .content {
            padding: 30px;
        }
        .meta-info {
            background: #151515;
            padding: 20px;
            border: 1px solid #2a2a2a;
            margin-bottom: 30px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
        }
        .meta-info div {
            margin: 8px 0;
            color: #c0c0c0;
        }
        .meta-info strong {
            color: #ff7e00;
        }
        .file-result {
            margin-bottom: 30px;
            border: 1px solid #2a2a2a;
            overflow: hidden;
        }
        .file-header {
            background: #1f1f1f;
            color: #ff7e00;
            padding: 15px 20px;
            font-weight: normal;
            font-size: 1em;
            border-bottom: 1px solid #2a2a2a;
        }
        .file-header::before {
            content: '▸ ';
        }
        .issue {
            padding: 20px;
            border-bottom: 1px solid #2a2a2a;
            background: #1a1a1a;
        }
        .issue:last-child {
            border-bottom: none;
        }
        .issue-header {
            display: flex;
            align-items: center;
            margin-bottom: 10px;
        }
        .severity-badge {
            padding: 4px 12px;
            border: 1px solid;
            font-size: 0.85em;
            font-weight: normal;
            text-transform: uppercase;
            margin-right: 12px;
        }
        .badge-critical {
            border-color: #ff3860;
            color: #ff3860;
            background: rgba(255, 56, 96, 0.1);
        }
        .badge-high {
            border-color: #ff7e00;
            color: #ff7e00;
            background: rgba(255, 126, 0, 0.1);
        }
        .badge-medium {
            border-color: #ffdd57;
            color: #ffdd57;
            background: rgba(255, 221, 87, 0.1);
        }
        .badge-low {
            border-color: #33ccff;
            color: #33ccff;
            background: rgba(51, 204, 255, 0.1);
        }
        .issue-description {
            font-size: 1em;
            color: #c0c0c0;
            margin-bottom: 10px;
            line-height: 1.5;
        }
        .issue-line {
            color: #666;
            font-size: 0.9em;
            margin-bottom: 10px;
        }
        .issue-line::before {
            content: '▸ ';
            color: #ff7e00;
        }
        .recommendation {
            background: rgba(255, 126, 0, 0.05);
            border-left: 3px solid #ff7e00;
            padding: 12px;
            margin-top: 10px;
            color: #c0c0c0;
        }
        .recommendation::before {
            content: '▸ ';
            color: #ff7e00;
            margin-right: 5px;
        }
        .no-issues {
            text-align: center;
            padding: 60px 20px;
            color: #33ccff;
        }
        .no-issues h2 {
            font-size: 2em;
            margin-bottom: 10px;
            font-weight: normal;
        }
        .no-issues h2::before {
            content: '▸ ';
            color: #ff7e00;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            border-top: 1px solid #2a2a2a;
            font-size: 0.85em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>SIDEKICK</h1>
            <p>AI Code Analysis</p>
        </div>

        <div class="summary">
            <div class="summary-card">
                <h3>Files Scanned</h3>
                <div class="number severity-total">{{.TotalFiles}}</div>
            </div>
            <div class="summary-card">
                <h3>Files With Findings</h3>
                <div class="number severity-high">{{.FilesWithIssues}}</div>
            </div>
        </div>

        <div class="content">
            <div class="meta-info">
                <div><strong>Scan Time:</strong> {{.Timestamp}}</div>
                <div><strong>Scan Path:</strong> {{.ScanPath}}</div>
                <div><strong>Model Used:</strong> {{.Model}}</div>
                <div><strong>Files Analyzed:</strong> {{.TotalFiles}}</div>
                <div><strong>Generated:</strong> {{.GenerationTime}}</div>
            </div>

            {{if eq .FilesWithIssues 0}}
            <div class="no-issues">
                <h2>No Issues Found</h2>
                <p>Analysis complete</p>
            </div>
            {{else}}
            {{range .Results}}
            {{if .HasIssues}}
            <div class="file-result">
                <div class="file-header">{{.FilePath}}</div>
                <div class="issue">
                    <div class="issue-description">{{.RawFindings | safeHTML}}</div>
                </div>
            </div>
            {{end}}
            {{end}}
            {{end}}
        </div>

        <div class="footer">
            <p>Generated by Sidekick v0.1.0 | Analysis powered by {{.Model}}</p>
        </div>
    </div>
</body>
</html>`

func GenerateHTML(results []scanner.ScanResult, scanPath, model string, totalFiles int, outputPath string) error {
	// Calculate statistics
	filesWithIssues := 0

	for _, result := range results {
		if result.HasIssues {
			filesWithIssues++
		}
	}

	// Create report data
	report := HTMLReport{
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
		ScanPath:        scanPath,
		Model:           model,
		TotalFiles:      totalFiles,
		FilesWithIssues: filesWithIssues,
		Results:         results,
		GenerationTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	// Parse template with custom functions
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template
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
