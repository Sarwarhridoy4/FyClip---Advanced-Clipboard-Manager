// File: main.go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/app"
	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/update"
)

// Version is set at build time - exported for UI access
var Version = "dev"

// BuildTime is set at build time - exported for UI access
var BuildTime = "unknown"

// GitHub repository info for update checking
const (
	githubOwner = "Sarwarhridoy4"
	githubRepo  = "FyClip---Advanced-Clipboard-Manager"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Print version info if requested
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		log.Printf("FyClip version %s (built: %s)", Version, BuildTime)
		os.Exit(0)
	}

	// Print help
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		log.Printf("FyClip - Advanced Clipboard Manager")
		log.Printf("Version: %s", Version)
		log.Printf("")
		log.Printf("Usage: fyclip [options]")
		log.Printf("  --version, -v            Show version")
		log.Printf("  --check-update          Check for updates")
		log.Printf("  --update                Download and install latest update")
		log.Printf("  --help, -h              Show help")
		os.Exit(0)
	}

	// Handle update commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--check-update":
			handleCheckUpdate()
			os.Exit(0)
		case "--update":
			handleUpdate()
			os.Exit(0)
		}
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Fatal panic: %v", r)
		}
	}()

	application := app.New()
	if application == nil {
		log.Fatal("Failed to create application")
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
		os.Exit(1)
	}
}

// handleCheckUpdate checks for updates and prints information
func handleCheckUpdate() {
	log.Printf("FyClip version %s (built: %s)", Version, BuildTime)
	log.Println("Checking for updates...")

	checker := update.NewChecker(githubOwner, githubRepo, Version)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	updateInfo, err := checker.CheckForUpdate(ctx)
	if err != nil {
		log.Printf("Error checking for updates: %v", err)
		log.Println("Make sure you have an internet connection.")
		os.Exit(1)
	}

	comparison := update.CompareVersions(updateInfo.LatestVersion, Version)
	if comparison > 0 {
		log.Printf("Update available: %s -> %s", Version, updateInfo.LatestVersion)
		log.Printf("Release notes: %s", updateInfo.ReleaseURL)
		log.Printf("Download: %s", updateInfo.DownloadURL)
		if updateInfo.IsPrerelease {
			log.Println("Note: This is a pre-release version")
		}
		log.Println("\nTo update, run: fyclip --update")
	} else {
		log.Printf("You are using the latest version: %s", Version)
	}
}

// handleUpdate downloads and installs the latest update
func handleUpdate() {
	log.Printf("FyClip version %s (built: %s)", Version, BuildTime)
	log.Println("Starting update process...")

	updater := update.NewAutoUpdater(githubOwner, githubRepo, Version)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 min timeout
	defer cancel()

	// Check and download
	log.Println("Checking for updates...")
	_, downloader, err := updater.CheckAndDownload(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

	if downloader == nil {
		log.Printf("You are already using the latest version: %s", Version)
		os.Exit(0)
	}

	log.Printf("Downloaded: %s", downloader.GetDownloadPath())

	// Install
	log.Println("Installing update...")
	installer := update.NewInstaller(downloader.GetDownloadPath(), "FyClip")
	if err := installer.Install(); err != nil {
		log.Printf("Installation error: %v", err)
		log.Println("You may need to install manually.")
		os.Exit(1)
	}

	log.Println("Update installed successfully!")
	log.Println("Please restart FyClip to use the new version.")
}