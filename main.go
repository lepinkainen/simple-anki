package main

import (
	"embed"
	"flag"
	"log"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	dbPath := flag.String("db", "flashcards.db", "Path to SQLite database")
	flag.Parse()

	// Initialize database
	if err := InitDB(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	// Setup routes
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/cards", CardsHandler)
	mux.HandleFunc("/api/cards/", CardHandler)
	mux.HandleFunc("/api/decks", DecksHandler)
	mux.HandleFunc("/api/review", ReviewHandler)
	mux.HandleFunc("/api/import", ImportHandler)

	// Serve static files from embedded filesystem
	mux.Handle("/", http.FileServer(http.FS(staticFiles)))

	log.Printf("Server starting on http://localhost:%s", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
