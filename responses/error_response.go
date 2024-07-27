package responses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, code int, mes string) {
	currentTime := time.Now().UTC()
	statusText := http.StatusText(code)

	c.JSON(http.StatusBadRequest, gin.H{
		"code":      code,
		"message":   statusText,
		"details":   mes,
		"timestamp": currentTime.Format(time.RFC3339),
	})
}
