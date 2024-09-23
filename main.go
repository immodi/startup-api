package main

import (
	"immodi/startup/handlers"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	r := handlers.MakeGinEngine()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	println("Currently Listening on http://localhost:8080....")
	r.Run(":" + port)
}
