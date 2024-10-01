package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"immodi/startup/responses"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

type AdminAuthResponse struct {
	Admin Admin  `json:"admin"`
	Token string `json:"token"`
}

type User struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type Admin struct {
	ID      string `json:"id"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Avatar  int    `json:"avatar"`
	Email   string `json:"email"`
}

type ErrorResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    ErrorData `json:"data"`
}

type ErrorData struct {
	Password ErrorDetail `json:"password"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func AuthenticateUserWithUsername(c echo.Context, app *pocketbase.PocketBase) error {
	var data User

	if err := c.Bind(&data); err != nil {
		return responses.PbErrorResponse(c, http.StatusUnprocessableEntity, "One or Both of keys called 'username' or 'password' is not present")
	}

	if data.Username == "" {
		return responses.PbErrorResponse(c, http.StatusUnprocessableEntity, "'username' is required")
	}

	if data.Password == "" {
		return responses.PbErrorResponse(c, http.StatusUnprocessableEntity, "'password' is required")
	}

	user, err := app.Dao().FindFirstRecordByData("users", "username", data.Username)
	if err != nil || !user.ValidatePassword(data.Password) {
		return responses.PbErrorResponse(c, http.StatusNotFound, "Invalid credentials")
	}

	return apis.RecordAuthResponse(app, c, user, nil)
}

func AuthenticateAdmin(c echo.Context, app *pocketbase.PocketBase) error {
	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request().Host)
	url := fmt.Sprintf("%s/api/admins/auth-with-password", baseURL)
	var response AdminAuthResponse

	data := struct {
		Identity string `json:"identity" form:"identity"`
		Password string `json:"password" form:"password"`
	}{}

	if err := c.Bind(&data); err != nil {
		fmt.Println("Failed to read request data", err)
		return responses.PbErrorResponse(c, http.StatusUnprocessableEntity, "One or Both of keys called 'identity' or 'password' is not present")
	}

	if data.Identity == "" {
		return responses.PbErrorResponse(c, http.StatusUnprocessableEntity, "'identity' is required")
	}

	if data.Password == "" {
		return responses.PbErrorResponse(c, http.StatusUnprocessableEntity, "'password' is required")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return responses.PbErrorResponse(c, http.StatusInternalServerError, "Error processing data, please try again")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return responses.PbErrorResponse(c, http.StatusInternalServerError, "Error processing data, please try again")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return responses.PbErrorResponse(c, http.StatusInternalServerError, "Error processing data, please try again")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return responses.PbErrorResponse(c, http.StatusInternalServerError, "Error processing data, please try again")
	}

	if resp.StatusCode > 199 && resp.StatusCode < 300 {
		if err := json.Unmarshal(body, &response); err == nil {
			return c.JSON(http.StatusAccepted, response)
		}

		fmt.Println("Error unmarshalling response JSON:", err)
	}

	return responses.PbErrorResponse(c, http.StatusInternalServerError, "Error processing data, please try again")
}
