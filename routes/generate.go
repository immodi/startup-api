package routes

import (
	"fmt"
	"immodi/startup/lib"
	"immodi/startup/responses"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v5"
)

type GinMessageRequest struct {
	Topic    string `json:"topic" binding:"required"`
	Template string `json:"template" binding:"required"`
}

type MessageRequest struct {
	Topic    string `json:"topic" form:"topic" binding:"required"`
	Template string `json:"template" form:"template" binding:"required"`
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

	lib.ParsePdfFile("data.html")

	c.JSON(http.StatusAccepted, gin.H{
		"response": response,
	})
}

func Generate(c echo.Context) error {
	var request MessageRequest

	if err := c.Bind(&request); err != nil {
		responses.PbErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return err
	}

	if request.Topic == "" {
		responses.PbErrorResponse(c, http.StatusBadRequest, "Missing required fields, 'topic'")
		return fmt.Errorf("missing required fields, 'topic' or 'template'")
	}

	message, usedTemplate := MessageBuilder(request.Topic, request.Template)
	response, err := GetAiResponse(message)

	if err != nil {
		responses.PbErrorResponse(c, http.StatusInternalServerError, "Couldn't contact AI model, please try again later")
		return err
	}

	err = lib.WriteResponseHTML(response, fmt.Sprintf("templates/%s.html", usedTemplate))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	lib.ParsePdfFile("data.html")

	// c.JSON(http.StatusAccepted, gin.H{
	// 	"response": response,
	// })
	c.File("data.pdf")

	return nil
}
