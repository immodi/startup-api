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
	"regexp"
	"strings"

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

func MessageBuilder(c echo.Context, app *pocketbase.PocketBase, topic string, templateName string, level int) (string, string) {
	//  htmlTemplate, _ := os.ReadFile("templates/snipets/template.html")
	htmlTemplate, err := repo.GetUserTemplateByName(c, app, templateName)
	styleTagString := extractStyleTagContent(htmlTemplate)

	if err != nil {
		htmlTemplateByteArray, _ := os.ReadFile("templates/snipets/document.html")
		htmlTemplate = string(htmlTemplateByteArray)
	}

	return fmt.Sprintf("Fill in the following HTML template %s with information about <topic>%s</topic>, please return the ONLY the requested content and no comments like {Let me know if you'd like me to add or modify anything!}, keep the vocabulary level at %d/10, and dont add any text to <br> tags", stripStyleTag(htmlTemplate), topic, level), styleTagString
}

func extractStyleTagContent(html string) string {
	// Case-insensitive regex pattern to match <style> tags and their content
	pattern := `(?i)<style\b[^>]*>([\s\S]*?)</style>`

	// Compile the regex pattern
	re := regexp.MustCompile(pattern)

	// Find the first match and extract the content inside the <style> tag
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		// Return the captured content inside the <style> tag
		return match[1]
	}

	// Return an empty string if no match is found
	return ""
}

func stripStyleTag(html string) string {
	// Case-insensitive regex pattern to match style tags and their content
	pattern := `(?i)<style\b[^>]*>[\s\S]*?</style>`

	// Compile the regex pattern
	re := regexp.MustCompile(pattern)

	// Replace all matches with empty string
	result := re.ReplaceAllString(html, "")

	// Trim any extra whitespace that might be left
	result = strings.TrimSpace(result)

	return result
}

func InsertStyleTag(html, styleContent string) string {
	// Find closing </div> tag (assuming it's the last one)
	lastDivIndex := strings.LastIndex(html, "</div>")

	if lastDivIndex == -1 {
		// If no </div> found, append to the end
		return fmt.Sprintf("%s\n<style>%s</style>", html, styleContent)
	}

	// Format the style tag with proper indentation
	styleTag := fmt.Sprintf("<style>%s</style>\n", styleContent)

	// Insert the style tag just before the last </div>
	result := html[:lastDivIndex] + styleTag + html[lastDivIndex:]

	return result
}
