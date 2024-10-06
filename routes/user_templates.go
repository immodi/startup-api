package routes

import (
	"immodi/startup/responses"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

func GetUserTemplates(c echo.Context, app *pocketbase.PocketBase) error {
	token := c.Request().Header.Get("Authorization")

	user, err := app.Dao().FindAuthRecordByToken(token, app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return responses.PbErrorResponse(c, http.StatusUnauthorized, "User doesn't exist for some reason")
	}

	templateIds := user.GetStringSlice("user_templates")
	templateNames := make([]string, 0)
	for _, templateId := range templateIds {
		templateName, _ := app.Dao().FindRecordById("templates", templateId)
		templateNames = append(templateNames, templateName.GetString("name"))
	}

	return c.JSON(http.StatusAccepted, map[string][]string{
		"templates": templateNames,
	})
}
