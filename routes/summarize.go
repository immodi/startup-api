package routes

import (
	"immodi/startup/ai"
	"immodi/startup/responses"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageRequest struct {
	Message string `json:"message" binding:"required"`
}

func Summarize(c *gin.Context) {
	var request MessageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "the field called 'message' is nonexistent")
		return
	}
	response, err := ai.AiResponse(request.Message)

	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Couldn't contact AI model, please try again later")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"summary": response,
	})
}
