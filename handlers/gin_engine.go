package handlers

import (
	"immodi/startup/routes"

	"github.com/gin-gonic/gin"
)

func MakeGinEngine() *gin.Engine {
	r := gin.Default()

	registerRoutes(r)

	return r
}

func registerRoutes(r *gin.Engine) *gin.Engine {
	r.POST("/", routes.Generate)

	return r
}
