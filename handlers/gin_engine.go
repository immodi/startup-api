package handlers

import (
	"immodi/startup/middlewares"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func GinEngine() *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())

	registerRoutes(r)
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
	return r
}

func registerRoutes(r *gin.Engine) *gin.Engine {
	// r.POST("/", routes.GinGenerate)

	return r
}
