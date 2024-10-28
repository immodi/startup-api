package routes

import (
	"fmt"
	"immodi/startup/lib"
	"immodi/startup/responses"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

type MessageRequest struct {
	Topic    string         `json:"topic" form:"topic" binding:"required"`
	Template string         `json:"template" form:"template" binding:"required"`
	Level    int            `json:"level" form:"level" binding:"required"`
	Data     map[string]any `json:"data" form:"data"`
}

func Generate(c echo.Context) error {
	var request MessageRequest
	var javascript string

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
		fmt.Println(err.Error())
		return err
	}

	filepath, err := lib.ParsePdfFile(lib.HtmlParserConfig{
		HtmlFileName:    "data.html",
		JavascriptToRun: javascript,
	})

	if err != nil {
		return responses.PbErrorResponse(c, 500, err.Error())
	}

	filename := strings.SplitAfter(filepath, "/")[1]
	return c.Attachment(filepath, filename)
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
