package interactive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pefman/sidekick/internal/ollama"
	"github.com/pefman/sidekick/internal/scanner"
)

func performScan(targetPath, modelName string, debug bool, scanType, customPrompt string) error {
	// Validate path
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	fmt.Printf("\n%s▸%s Scanning: %s\n", orange, reset, targetPath)
	fmt.Printf("%s▸%s Model: %s\n\n", orange, reset, modelName)

	// Initialize Ollama client
	client := ollama.NewClient("http://localhost:11434")

	// Check if model is available
	if err := client.CheckModel(modelName); err != nil {
		return fmt.Errorf("model check failed: %w\nMake sure Ollama is running and the model is installed", err)
	}

	// Initialize scanner
	s := scanner.NewScanner(client, modelName, debug, scanType, customPrompt)
	defer s.Close()

	// Collect files
	var files []string
	if info.IsDir() {
		files, err = collectFiles(targetPath)
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}
	} else {
		files = []string{targetPath}
	}

	if len(files) == 0 {
		fmt.Println("No files to scan")
		return nil
	}

	fmt.Printf("%s▸%s Found %d files to analyze\n\n", orange, reset, len(files))

	// Scan files
	results, err := s.ScanFiles(files)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display results
	displayResults(results)

	return nil
}

func collectFiles(root string) ([]string, error) {
	var files []string
	extensions := []string{".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".rs", ".rb", ".php"}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" ||
				name == "vendor" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip sensitive files
		name := info.Name()
		sensitiveFiles := []string{".env", ".env.local", ".env.production", "id_rsa", "id_ed25519", ".pem", ".key", ".pfx", ".p12"}
		for _, sensitive := range sensitiveFiles {
			if name == sensitive || strings.HasSuffix(name, sensitive) {
				return nil // Skip this file
			}
		}

		ext := filepath.Ext(path)
		for _, validExt := range extensions {
			if ext == validExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

func displayResults(results []scanner.ScanResult) {
	filesWithIssues := 0

	for _, result := range results {
		if result.HasIssues {
			filesWithIssues++
			fmt.Printf("\n%s━━━ %s ━━━%s\n", orange, filepath.Base(result.FilePath), reset)
			fmt.Println(result.RawFindings)
			fmt.Println()
		}
	}

	// Summary
	fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", orange, reset)
	fmt.Printf("%s📊 Scan Summary%s\n", orange, reset)
	fmt.Printf("   Files scanned: %d\n", len(results))
	fmt.Printf("   Files with findings: %d\n", filesWithIssues)
	if filesWithIssues == 0 {
		fmt.Printf("   %s✓%s No issues detected!\n", cyan, reset)
	}
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", orange, reset)
}
