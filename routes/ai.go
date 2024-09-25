package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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
		log.Fatalf("Error loading .env file")
		return "", err
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println(".")
		return "Error using the service, please try again later", err
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

func MessageBuilder(topic string, templateId string) (string, string) {
	htmlTemplate, err := os.ReadFile(fmt.Sprintf("templates/snipets/%s.html", templateId))

	if err != nil {
		return MessageBuilder(topic, "document")
	}

	return fmt.Sprintf("Fill in the following HTML template with information about %s: %s", topic, htmlTemplate), templateId
}
