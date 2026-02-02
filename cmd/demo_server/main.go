package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// serve from the "web" directory relative to project root
	rootDir, _ := os.Getwd()
	webDir := filepath.Join(rootDir, "web")

	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	port := "8080"
	yellow := "\033[33m"
	reset := "\033[0m"
	
	fmt.Println(yellow + "=================================================================" + reset)
	fmt.Println(yellow + "   🛡️  TALOS SALES DEMO SERVER" + reset)
	fmt.Println(yellow + "=================================================================" + reset)
	fmt.Printf("📂 Serving content from: %s\n", webDir)
	fmt.Printf("🚀 Open Landing Page:    http://localhost:%s/landing.html\n", port)
	fmt.Printf("✨ Open Dashboard:       http://localhost:%s/index.html\n", port)
	fmt.Println(yellow + "=================================================================" + reset)
	fmt.Println("Press Ctrl+C to stop the server...")

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
