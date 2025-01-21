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

	if request.Data != nil {
		javascript = jsInjectionScript(&request.Data)
	}

	if request.Topic == "" {
		return responses.PbErrorResponse(c, http.StatusBadRequest, "Missing required fields, 'topic' or 'template'")
	}

	message, styleTag := MessageBuilder(c, app, request.Topic, request.Template, request.Level)

	htmlData := lib.HtmlFileData{
		UserName:       lib.GenerateUniqueString(user.Username()),
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
	}, htmlData.UserName)

	if err != nil {
		return responses.PbErrorResponse(c, 500, err.Error())
	}

	go lib.StoreFile(app, user.Id, filepath)

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
		UserName:       lib.GenerateUniqueString(htmlData.UserName),
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
	sb.WriteString(`() => {`)

	for key, value := range *data {
		if valueString, ok := value.(string); ok {
			sb.WriteString(fmt.Sprintf(`
				try {
					document.querySelector("#%s").innerHTML = %s;
				} catch (error) {}
			`, key, strconv.Quote(valueString)))
		}
	}

	sb.WriteString(`}`)
	return sb.String()
}
