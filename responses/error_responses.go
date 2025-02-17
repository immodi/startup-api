package responses

import (
	"fmt"
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

func NotAllowed(c *gin.Context) {
	ErrorResponse(c, http.StatusMethodNotAllowed, fmt.Sprintf("[%s] Method Not Allowed", c.Request.Method))
}
