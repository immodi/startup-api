package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"immodi/startup/repo"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GetAiResponse(message string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, will try to see the system ENV variables")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalf("Error loading env var")
		return "Error using the service, please try again later", fmt.Errorf("error loading env file")
	}

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": message,
			},
		},
		"model": "llama3-70b-8192",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var groqResp GroqResponse
	err = json.Unmarshal(body, &groqResp)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return groqResp.Choices[0].Message.Content, nil
}

func MessageBuilder(c echo.Context, app *pocketbase.PocketBase, topic string, templateName string, level int) string {
	//  htmlTemplate, _ := os.ReadFile("templates/snipets/template.html")
	htmlTemplate, err := repo.GetUserTemplateByName(c, app, templateName)
	if err != nil {
		htmlTemplateByteArray, _ := os.ReadFile("templates/snipets/document.html")
		htmlTemplate = string(htmlTemplateByteArray)
	}

	return fmt.Sprintf("Fill in the following HTML template %s with information about <topic>%s</topic>, please return the ONLY the requested content and no comments like {Let me know if you'd like me to add or modify anything!}, keep the vocabulary level at %d/10", htmlTemplate, topic, level)
}
