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
)

type MessageRequest struct {
	Topic    string         `json:"topic" form:"topic" binding:"required"`
	Template string         `json:"template" form:"template" binding:"required"`
	Level    int            `json:"level" form:"level" binding:"required"`
	Data     map[string]any `json:"data" form:"data"`
}

func Generate(c echo.Context, app *pocketbase.PocketBase) error {
	var request MessageRequest
	var javascript string

	token := c.Request().Header.Get("Authorization")

	if err := c.Bind(&request); err != nil {
		return responses.PbErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	user, err := app.Dao().FindAuthRecordByToken(token, app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return err
	}

	if userTokens := user.GetInt("tokens"); userTokens < 1 {
		return responses.PbErrorResponse(c, 400, "not enough tokens, smh")
	}

	go lib.ClearDirectory(fmt.Sprintf("pdfs/%s", user.Username()))

	if request.Data != nil {
		javascript = jsInjectionScript(&request.Data)
	}

	if request.Topic == "" {
		return responses.PbErrorResponse(c, http.StatusBadRequest, "Missing required fields, 'topic' or 'template'")
	}

	message, styleTag := MessageBuilder(c, app, request.Topic, request.Template, request.Level)

	htmlData := lib.HtmlFileData{
		Username:       user.Username(),
		UserBlobName:   lib.GenerateUniqueString(user.Username()),
		TemplateName:   request.Template,
		HtmlData:       "",
		StyleTag:       styleTag,
		InsertStyleTag: InsertStyleTag,
	}

	htmlFilePath, err := getAIResponseAndWriteHTMl(c, app, &htmlData, message, 3)
	if err != nil {
		return responses.PbErrorResponse(c, 500, err.Error())
	}

	filepath, err := lib.ParsePdfFile(c, app, lib.HtmlParserConfig{
		TemplateName:    htmlFilePath,
		JavascriptToRun: javascript,
		OutputFileName:  htmlData.UserBlobName,
		Username:        user.Username(),
	})

	if err != nil {
		return responses.PbErrorResponse(c, 500, err.Error())
	}

	go lib.StoreFile(app, user.Id, filepath)
	go lib.ClearDirectory(fmt.Sprintf("user_data/%s", user.Username()))

	filename := strings.SplitAfter(filepath, "/")[1]
	err = c.Attachment(filepath, filename)
	if err != nil {
		println(fmt.Sprintf("FileResponse Error: %s \n", err.Error()))
	}

	return err
}

func getAIResponseAndWriteHTMl(c echo.Context, app *pocketbase.PocketBase, htmlData *lib.HtmlFileData, message string, tries int) (string, error) {
	response, err := GetAiResponse(message)

	if err != nil {
		if tries < 1 {
			return "", fmt.Errorf("error in the ai serivce, please try again later")
		}
		return getAIResponseAndWriteHTMl(c, app, htmlData, message, tries-1)
	}

	htmlFilePath, err := lib.WriteResponseHTML(c, app, &lib.HtmlFileData{
		Username:       htmlData.Username,
		UserBlobName:   htmlData.UserBlobName,
		TemplateName:   htmlData.TemplateName,
		HtmlData:       response,
		StyleTag:       htmlData.StyleTag,
		InsertStyleTag: htmlData.InsertStyleTag,
	})
	if err != nil {
		return "", err
	}

	return htmlFilePath, nil
}

func jsInjectionScript(data *map[string]any) string {
	var sb strings.Builder

	// Start the event listener for DOMContentLoaded
	sb.WriteString(`
		document.addEventListener("DOMContentLoaded", function() {
	`)

	// Iterate over the map and generate JS code
	for key, value := range *data {
		if valueString, ok := value.(string); ok {
			sb.WriteString(fmt.Sprintf(`
				try {
					document.querySelector("#%s").innerHTML = %s;
				} catch (error) {
					console.error("Error updating element #%s:", error);
				}
			`, key, strconv.Quote(valueString), key))
		}
	}

	// End the event listener
	sb.WriteString(`
		});
	`)

	return sb.String()
}
