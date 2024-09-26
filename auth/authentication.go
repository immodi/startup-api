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
	data := struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}{}
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Failed to read request data", err)
	}

	record, err := app.Dao().FindFirstRecordByData("users", "username", data.Username)
	if err != nil || !record.ValidatePassword(data.Password) {
		return apis.NewBadRequestError("Invalid credentials", err)
	}

	return apis.RecordAuthResponse(app, c, record, nil)
}

func AuthenticateAdmin(c echo.Context, app *pocketbase.PocketBase) error {
	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request().Host)
	url := fmt.Sprintf("%s/api/admins/auth-with-password", baseURL)
	var response AdminAuthResponse
	var errorResponse ErrorResponse

	data := struct {
		Identity string `json:"identity" form:"identity"`
		Password string `json:"password" form:"password"`
	}{}

	if err := c.Bind(&data); err != nil {
		fmt.Println("Failed to read request data", err)
		return apis.NewBadRequestError("Failed to read request data", err)
	}

	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return apis.NewBadRequestError("Error marshalling JSON:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return apis.NewBadRequestError("Error creating request:", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return apis.NewBadRequestError("Error sending request:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return apis.NewBadRequestError("Error reading response body", err)
	}

	if resp.StatusCode > 199 && resp.StatusCode < 300 {
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Println("Error unmarshalling response JSON:", err)
			return apis.NewBadRequestError("Error unmarshalling response JSON", err)
		}
		return c.JSON(200, response)
	}

	if resp.StatusCode > 399 && resp.StatusCode < 500 {
		if err = json.Unmarshal(body, &errorResponse); err != nil {
			fmt.Println("Error unmarshalling response JSON:", err)
			return apis.NewBadRequestError("Error unmarshalling response JSON", err)
		}
		return c.JSON(400, errorResponse)
	}

	responses.PbErrorResponse(c, 500, "internal server error")
	return nil
}
