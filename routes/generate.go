package routes

import (
	"fmt"
	"immodi/startup/lib"
	"immodi/startup/responses"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageRequest struct {
	Message string `json:"message" binding:"required"`
}

func Generate(c *gin.Context) {
	var request MessageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "the field called 'message' is nonexistent")
		return
	}
	response, err := GetAiResponse(request.Message)

	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Couldn't contact AI model, please try again later")
		return
	}

	err = lib.WriteResponseHTML(response)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	lib.ParsePdfFile("data.html")

	// c.JSON(http.StatusAccepted, gin.H{
	// 	"response": response,
	// })

	c.FileAttachment("data.pdf", "data.pdf")
}
