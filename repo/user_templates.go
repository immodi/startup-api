package repo

import (
	"fmt"
	"immodi/startup/responses"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

func GetUserTemplates(c echo.Context, app *pocketbase.PocketBase) error {
	token := c.Request().Header.Get("Authorization")

	if token == "" {
		defaultTemplates := []*models.Record{}

		err := app.Dao().RecordQuery("templates").Limit(3).All(&defaultTemplates)
		if err != nil {
			return responses.PbErrorResponse(c, http.StatusUnauthorized, "No Templates avaliable")
		}

		templates := make([][]string, 0)
		for _, template := range defaultTemplates {
			templates = append(templates, []string{template.Id, template.GetString("name")})
		}

		return c.JSON(http.StatusAccepted, templates)
	}

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
		return "", fmt.Errorf("user doesn't exist for some reason")
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

	return "", fmt.Errorf("no user template found")
}

func GetTemplateSourceContent(c echo.Context, app *pocketbase.PocketBase, userTemplateName string) (string, error) {
	token := c.Request().Header.Get("Authorization")

	user, err := app.Dao().FindAuthRecordByToken(token, app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return "", fmt.Errorf("user doesn't exist for some reason")
	}

	templateIds := user.GetStringSlice("user_templates")
	for _, templateId := range templateIds {

		template, err := app.Dao().FindRecordById("templates", templateId)
		if err != nil {
			break
		}

		if template.Get("name") == userTemplateName {
			templateSourceId := template.GetString("source_template")
			templateSource, err := app.Dao().FindRecordById("sources", templateSourceId)
			if err != nil {
				return "", fmt.Errorf("template not found")
			}

			if source := templateSource.GetString("content"); source != "" {
				log.Println(source)
				return source, nil
			}

		}
	}
	content, _ := os.ReadFile("lib/templates/document.html")

	return string(content), nil
}
