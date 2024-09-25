package routes

import (
	"fmt"
	"immodi/startup/lib"
	"immodi/startup/responses"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageRequest struct {
	Topic    string `json:"topic" binding:"required"`
	Template string `json:"template" binding:"required"`
}

func Generate(c *gin.Context) {
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

	// c.FileAttachment("data.pdf", "data.pdf")
}
