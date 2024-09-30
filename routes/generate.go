package routes

import (
	"fmt"
	"immodi/startup/lib"
	"immodi/startup/responses"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v5"
)

type GinMessageRequest struct {
	Topic    string `json:"topic" binding:"required"`
	Template string `json:"template" binding:"required"`
}

type MessageRequest struct {
	Topic    string         `json:"topic" form:"topic" binding:"required"`
	Template string         `json:"template" form:"template" binding:"required"`
	Data     map[string]any `json:"data" form:"data"`
}

func GinGenerate(c *gin.Context) {
	var request MessageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "the fields called 'message' and 'template' are nonexistent")
		return
	}

	message, usedTemplate := MessageBuilder(request.Topic, request.Template)
	response, err := GetAiResponse(message)

	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Couldn't contact AI model, please try again later")
		return
	}

	err = lib.WriteResponseHTML(response, fmt.Sprintf("templates/%s.html", usedTemplate))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	lib.ParsePdfFile(lib.HtmlParserConfig{
		HtmlFileName: "data.html",
		JavascriptToRun: `() => {
			// Example: Modify the DOM, add styles, or perform any action
			document.body.style.backgroundColor = "lightblue";
		}`,
	})

	c.JSON(http.StatusAccepted, gin.H{
		"response": response,
	})
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

	message, usedTemplate := MessageBuilder(request.Topic, request.Template)
	response, err := GetAiResponse(message)

	if err != nil {
		return responses.PbErrorResponse(c, http.StatusInternalServerError, "Couldn't contact AI model, please try again later")
	}

	err = lib.WriteResponseHTML(response, fmt.Sprintf("templates/%s.html", usedTemplate))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	lib.ParsePdfFile(lib.HtmlParserConfig{
		HtmlFileName:    "data.html",
		JavascriptToRun: javascript,
	})

	// c.JSON(http.StatusAccepted, gin.H{
	// 	"response": response,
	// })

	c.File("data.pdf")

	return nil
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
