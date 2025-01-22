package repo

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
	templates := make([][]string, 0)
	for _, templateId := range templateIds {
		templateName, _ := app.Dao().FindRecordById("templates", templateId)
		templates = append(templates, []string{templateId, templateName.GetString("name")})
	}

	return c.JSON(http.StatusAccepted, templates)
}

func GetUserTemplateByName(c echo.Context, app *pocketbase.PocketBase, userTemplateName string) (string, error) {
	token := c.Request().Header.Get("Authorization")

	user, err := app.Dao().FindAuthRecordByToken(token, app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return "", responses.PbErrorResponse(c, http.StatusUnauthorized, "User doesn't exist for some reason")
	}

	templateIds := user.GetStringSlice("user_templates")
	for _, templateId := range templateIds {
		template, err := app.Dao().FindRecordById("templates", templateId)
		if err != nil {
			break
		}

		templateName := template.GetString("name")
		if userTemplateName == templateName {
			template, _ := app.Dao().FindRecordById("templates", templateId)
			return template.GetString("html"), nil
		}
	}

	return "", responses.PbErrorResponse(c, http.StatusNotFound, "Template not found")
}

func GetTemplateSourceContent(c echo.Context, app *pocketbase.PocketBase, userTemplateName string) (string, error) {
	token := c.Request().Header.Get("Authorization")

	user, err := app.Dao().FindAuthRecordByToken(token, app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return "", responses.PbErrorResponse(c, http.StatusUnauthorized, "User doesn't exist for some reason")
	}

	templateIds := user.GetStringSlice("user_templates")
	for _, templateId := range templateIds {
		template, err := app.Dao().FindRecordById("templates", templateId)
		if err != nil {
			break
		}

		templateSourceId := template.GetString("source_template")
		templateSource, err := app.Dao().FindRecordById("sources", templateSourceId)
		if err != nil {
			return "", responses.PbErrorResponse(c, http.StatusNotFound, "Template not found")
		}

		if source := templateSource.GetString("content"); source != "" {
			return source, nil
		}
	}

	return "", responses.PbErrorResponse(c, http.StatusNotFound, "Template not found")
}
