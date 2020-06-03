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

	//scraper := engine.NewScraper()
	//scraper.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("PORT env var not found, using 8080 as default")
	}

	err := http.ListenAndServe(":"+port, server)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
		os.Exit(1)
	}
}
