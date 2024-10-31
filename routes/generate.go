package routes

import (
	"fmt"
	"immodi/startup/lib"
	"immodi/startup/responses"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type MessageRequest struct {
	Topic    string         `json:"topic" form:"topic" binding:"required"`
	Template string         `json:"template" form:"template" binding:"required"`
	Level    int            `json:"level" form:"level" binding:"required"`
	Data     map[string]any `json:"data" form:"data"`
}

type UserFile struct {
	filepath string
	token    string
}

func Generate(c echo.Context, app *pocketbase.PocketBase) error {
	var request MessageRequest
	var javascript string

	token := c.Request().Header.Get("Authorization")

	if err := c.Bind(&request); err != nil {
		return responses.PbErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if request.Data != nil {
		javascript = jsInjectionScript(&request.Data)
	}

	if request.Topic == "" {
		return responses.PbErrorResponse(c, http.StatusBadRequest, "Missing required fields, 'topic' or 'template'")
	}

	message, usedTemplate := MessageBuilder(request.Topic, request.Template, request.Level)
	response, err := GetAiResponse(message)

	if err != nil {
		return responses.PbErrorResponse(c, http.StatusInternalServerError, "Couldn't contact AI model, please try again later")
	}

	err = lib.WriteResponseHTML(response, fmt.Sprintf("templates/%s.html", usedTemplate))
	if err != nil {
		return err
	}

	filepath, err := lib.ParsePdfFile(lib.HtmlParserConfig{
		HtmlFileName:    "data.html",
		JavascriptToRun: javascript,
	})

	if err != nil {
		return responses.PbErrorResponse(c, 500, err.Error())
	}

	go storeUserFile(app, UserFile{
		filepath: filepath,
		token:    token,
	})

	filename := strings.SplitAfter(filepath, "/")[1]
	err = c.Attachment(filepath, filename)
	if err != nil {
		println(fmt.Sprintf("FileResponse Error: %s \n", err.Error()))
	}

	return err
}

func jsInjectionScript(data *map[string]any) string {
	var sb strings.Builder
	sb.WriteString(`() => {`)

	for key, value := range *data {
		if valueString, ok := value.(string); ok {
			sb.WriteString(fmt.Sprintf(`
				try {
					document.querySelector(".%s").innerHTML = %s;
				} catch (e) {
                    console.log(e)
                }
			`, key, strconv.Quote(valueString)))
		}
	}

	sb.WriteString(`}`)
	return sb.String()
}

func storeFile(app *pocketbase.PocketBase, filepath string) (string, error) {
	collection, err := app.Dao().FindCollectionByNameOrId("files")
	if err != nil {
		return "", err
	}

	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)

	f1, err := filesystem.NewFileFromPath(filepath)
	if err != nil {
		return "", err
	}

	form.AddFiles("file", f1)

	if err := form.Submit(); err != nil {
		return "", err
	}

	return record.Id, nil
}

func updateUserRecord(app *pocketbase.PocketBase, user *models.Record, fileId string) error {
	filesId := user.GetStringSlice("user_files")
	form := forms.NewRecordUpsert(app, user)

	form.LoadData(map[string]any{
		"user_files": append(filesId, fileId),
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}

func storeUserFile(app *pocketbase.PocketBase, userData UserFile) error {
	user, err := app.Dao().FindAuthRecordByToken(userData.token, app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return err
	}

	fileId, err := storeFile(app, userData.filepath)
	if err != nil {
		return err
	}

	err = updateUserRecord(app, user, fileId)
	if err != nil {
		return err
	}

	return nil
}
