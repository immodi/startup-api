package responses

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
)

func PbErrorResponse(c echo.Context, code int, mes string) *apis.ApiError {
	currentTime := time.Now().UTC()
	statusText := http.StatusText(code)
	err := apis.NewApiError(
		code,
		mes,
		nil,
	)

	c.JSON(http.StatusBadRequest, map[string]any{
		"code":      err.Code,
		"message":   statusText,
		"details":   err.Message,
		"timestamp": currentTime.Format(time.RFC3339),
	})

	return err
}
