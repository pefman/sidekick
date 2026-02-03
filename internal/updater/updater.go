package updater

import (
	"context"
	"fmt"
	"os"

	"github.com/creativeprojects/go-selfupdate"
)

const (
	repo = "pefman/sidekick"
)

// Version is set at build time via -ldflags
var Version = "dev"

// CheckForUpdate checks if a new version is available
func CheckForUpdate() (*selfupdate.Release, bool, error) {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug(repo))
	if err != nil {
		return nil, false, fmt.Errorf("error checking for updates: %w", err)
	}

	if !found {
		return nil, false, fmt.Errorf("no releases found")
	}

	// Compare versions
	v := Version
	if v == "dev" {
		v = "v0.0.0" // Treat dev as very old version
	}

	if latest.GreaterThan(v) {
		return latest, true, nil
	}

	return latest, false, nil
}

// Update downloads and installs the latest version
func Update() error {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("could not create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.ParseSlug(repo))
	if err != nil {
		return fmt.Errorf("error checking for updates: %w", err)
	}
	if !found {
		return fmt.Errorf("no releases found")
	}

	// Compare versions
	v := Version
	if v == "dev" {
		v = "v0.0.0"
	}

	if !latest.GreaterThan(v) {
		return fmt.Errorf("already running latest version: %s", Version)
	}

	fmt.Printf("📦 Downloading version %s...\n", latest.Version())

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	// Check if we need sudo for the update
	info, err := os.Stat(exe)
	if err != nil {
		return fmt.Errorf("could not stat executable: %w", err)
	}

	// If not writable, we need sudo
	if info.Mode().Perm()&0200 == 0 {
		return fmt.Errorf("executable is not writable. Please run with sudo or as root")
	}

	err = updater.UpdateTo(context.Background(), latest, exe)
	if err != nil {
		return fmt.Errorf("error installing update: %w", err)
	}

	fmt.Printf("✅ Updated to version %s\n", latest.Version())
	fmt.Println("Please restart sidekick to use the new version.")

	return nil
}
