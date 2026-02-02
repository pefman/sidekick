package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pefman/sidekick/internal/config"
	"github.com/pefman/sidekick/internal/ollama"
	"github.com/pefman/sidekick/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	targetPath string
	modelName  string
	debug      bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan codebase for security issues",
	Long:  `Scan your codebase for security vulnerabilities using local LLM via Ollama.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScan,
}

func init() {
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = config.GetDefault()
	}

	scanCmd.Flags().StringVarP(&modelName, "model", "m", cfg.DefaultModel, "Ollama model to use")
	scanCmd.Flags().BoolVarP(&debug, "debug", "d", cfg.Debug, "Enable debug logging to file")
}

func runScan(cmd *cobra.Command, args []string) error {
	// Determine target path
	if len(args) > 0 {
		targetPath = args[0]
	} else {
		var err error
		targetPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Validate and clean path to prevent directory traversal
	var err error
	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	targetPath = filepath.Clean(targetPath)

	// Verify path exists and is accessible
	if _, err := os.Stat(targetPath); err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	// Validate path exists
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	fmt.Printf("🔍 Scanning: %s\n", targetPath)
	fmt.Printf("🤖 Using model: %s\n\n", modelName)

	// Initialize Ollama client
	client := ollama.NewClient("http://localhost:11434")

	// Check if model is available
	if err := client.CheckModel(modelName); err != nil {
		return fmt.Errorf("model check failed: %w\nMake sure Ollama is running and the model is installed", err)
	}

	// Initialize scanner
	s := scanner.NewScanner(client, modelName, debug, "security", "")
	defer s.Close()

	// Scan files
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

	fmt.Printf("📁 Found %d files to analyze\n\n", len(files))

	// Scan each file
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

		// Skip hidden directories and common ignore patterns
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

		// Check if file has relevant extension
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
			fmt.Printf("\n\033[38;5;208m━━━ %s ━━━\033[0m\n", filepath.Base(result.FilePath))
			fmt.Println(result.RawFindings)
			fmt.Println()
		}
	}

	// Summary
	fmt.Println("\n\033[38;5;208m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m")
	fmt.Printf("\033[38;5;208m📊 Scan Summary\033[0m\n")
	fmt.Printf("   Files scanned: %d\n", len(results))
	fmt.Printf("   Files with findings: %d\n", filesWithIssues)
	if filesWithIssues == 0 {
		fmt.Println("   \033[38;5;82m✓\033[0m No issues detected!")
	}
	fmt.Println("\033[38;5;208m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m")
}
