// File: main.go
package main

import (
	"log"
	"os"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/app"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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