// File: main.go
package main

import (
	"log"
	"os"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/app"
)

// Version is set at build time
var version = "dev"

// buildTime is set at build time
var buildTime = "unknown"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Print version info if requested
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		log.Printf("FyClip version %s (built: %s)", version, buildTime)
		os.Exit(0)
	}

	// Print help
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		log.Printf("FyClip - Advanced Clipboard Manager")
		log.Printf("Version: %s", version)
		log.Printf("")
		log.Printf("Usage: fyclip [options]")
		log.Printf("  --version, -v  Show version")
		log.Printf("  --help, -h     Show help")
		os.Exit(0)
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