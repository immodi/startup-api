package responses

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

func PbErrorResponse(c echo.Context, code int, mes string) {
	currentTime := time.Now().UTC()
	statusText := http.StatusText(code)

	c.JSON(http.StatusBadRequest, map[string]any{
		"code":      code,
		"message":   statusText,
		"details":   mes,
		"timestamp": currentTime.Format(time.RFC3339),
	})
}
