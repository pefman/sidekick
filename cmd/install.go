package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install sidekick to /usr/local/sbin",
	Long:  `Install the sidekick binary to /usr/local/sbin for system-wide access.`,
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks and clean path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}
	execPath = filepath.Clean(execPath)

	// Validate source file exists and is a regular file
	sourceInfo, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("failed to access source file: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", execPath)
	}

	// Use fixed, absolute target path (not user-controllable)
	targetPath := "/usr/local/sbin/sidekick"
	targetPath = filepath.Clean(targetPath)

	fmt.Printf("📦 Installing sidekick...\n")
	fmt.Printf("   Source: %s\n", execPath)
	fmt.Printf("   Target: %s\n\n", targetPath)

	// Check if target directory exists, create if not
	targetDir := filepath.Dir(targetPath)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", targetDir)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory (try with sudo): %w", err)
		}
	}

	// Copy binary
	if err := copyFile(execPath, targetPath); err != nil {
		return fmt.Errorf("failed to copy binary (try with sudo): %w", err)
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Println("✅ Installation successful!")
	fmt.Println("\nYou can now run 'sidekick' from anywhere!")
	fmt.Println("\nExample commands:")
	fmt.Println("  sidekick scan")
	fmt.Println("  sidekick scan /path/to/project")
	fmt.Println("  sidekick scan --model codellama")

	return nil
}

func copyFile(src, dst string) error {
	// Validate and clean paths
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	// Verify source exists and is a regular file
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot access source: %w", err)
	}
	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file")
	}

	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}
