package main

import (
	"aniapi-go/api"
	"aniapi-go/database"
	"aniapi-go/engine"
	"log"
	"net/http"
	"os"
)

func main() {
	server := engine.NewServer()

	server.Handle("/api/.*", api.Router)

	database.Init()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("PORT env var not found, using 8080 as default")
	}

	go engine.StartQueue()

	scraper := engine.NewScraper()
	go scraper.Start()

	err := http.ListenAndServe(":"+port, server)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
		os.Exit(1)
	}
}
