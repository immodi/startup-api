package lib

import (
	"embed"
	"fmt"
	"html/template"
	"immodi/startup/repo"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/russross/blackfriday/v2"
)

//go:embed templates/*
var Templates embed.FS

type Tag struct {
	Index     int
	Type      bool //`opening:1 - closing:0`
	TagLength int
}

type HtmlFileData struct {
	Username       string
	UserBlobName   string
	TemplateName   string
	HtmlData       string
	StyleTag       string
	InsertStyleTag func(string, string) string
}

func WriteResponseHTML(c echo.Context, app *pocketbase.PocketBase, htmlData *HtmlFileData) (string, error) {
	htmlData.HtmlData = htmlData.InsertStyleTag(RemoveTrailingFreeText(htmlData.HtmlData), htmlData.StyleTag)

	templateHtml, err := getSourceHtmlContent(c, app, htmlData.TemplateName)
	if err != nil {
		return "", err
	}

	// Create full directory path first
	dirPath := fmt.Sprintf("user_data/%s", htmlData.Username)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", err
	}

	// Then create the file
	fileName := fmt.Sprintf("%s/%s.html", dirPath, htmlData.UserBlobName)
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new template and parse the template string
	tmpl, err := template.New("page").Parse(templateHtml)
	if err != nil {
		return "", err
	}

	// Write the template output to the file
	err = tmpl.Execute(file, struct {
		Content template.HTML
	}{
		Content: template.HTML(htmlData.HtmlData),
	})
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func RemoveTrailingFreeText(htmlData string) string {
	index := strings.IndexRune(htmlData, '<')

	if index == -1 {
		return htmlData
	}

	return htmlData[index:]
}

func SanatizeHtml(badHtml string) string {
	// badHtml := badHtml() + " "
	badIndex := make([]Tag, 0)

	for index, char := range badHtml {
		if char == '<' {
			if badHtml[index+1] != '/' {
				_, supposedClosingIndex, tagLength := getTag(&badHtml, index, "")
				if badHtml[supposedClosingIndex] != '>' {
					badIndex = append(badIndex, Tag{
						Index:     index + 2,
						Type:      true,
						TagLength: tagLength,
					})
				}
			} else {
				char := rune(badHtml[index+2])
				tagLength := getTagLength(char)
				// println(string(badHtml[index+2+tagLength]))
				if badHtml[index+2+tagLength] != '>' {
					badIndex = append(badIndex, Tag{
						Index:     index + tagLength + 1,
						Type:      false,
						TagLength: tagLength,
					})
				}
			}
		}
	}

	for shiftIndex, tag := range badIndex {
		badHtml = insertAtIndex(badHtml, tag.Index+1+shiftIndex, '>')
	}

	return badHtml
}

func ConvertMarkdownToHTML(markdown string) string {
	// Convert Markdown to HTML
	html := blackfriday.Run([]byte(markdown))
	return string(html)
}

func getTag(refString *string, index int, tag string) (string, int, int) {
	t := tag
	stringArray := []byte(*refString)

	if stringArray[index+1] != '>' {
		t += string(stringArray[index+1])
		return getTag(refString, index+1, t)
	}

	tagCharLength := getTagLength(rune(tag[0]))
	if tagCharLength == 0 {
		tagCharLength = 2
	}

	return tag, index - (len(tag) - 1) + tagCharLength, tagCharLength
}

func getTagLength(tagRune rune) int {
	// Define a map to associate constant values with their names
	constNames := map[rune]int{
		'h': 2,
		'u': 2,
		'p': 1,
		'l': 2,
		'd': 3,
		's': 7,
	}

	// Lookup the value in the map
	if len, exists := constNames[tagRune]; exists {
		return len
	}
	return 0
}

func insertAtIndex(original string, index int, newContent rune) string {
	// Ensure the index is within bounds
	if index < 0 || index > len(original) {
		return original // Index out of bounds
	}

	// Split the original string into two parts
	before := original[:index]
	after := original[index:]

	// Concatenate the parts with the new content in between
	return before + string(newContent) + after
}

func getSourceHtmlContent(c echo.Context, app *pocketbase.PocketBase, userTemplateName string) (string, error) {
	htmlData, err := repo.GetTemplateSourceContent(c, app, userTemplateName)

	if err != nil {
		localTemplates := make(map[string]string)
		localTemplates["document"] = "document.html"
		localTemplates["report"] = "report.html"
		localTemplates["paragraph"] = "paragraph.html"
		localTemplates["template"] = "template.html"

		var templatePath string
		if _, ok := localTemplates[userTemplateName]; ok {
			templatePath = localTemplates[userTemplateName]
		} else {
			templatePath = localTemplates["template"]
		}

		templateHtml, err := Templates.ReadFile(fmt.Sprintf("templates/%s", templatePath))
		if err != nil {
			return "", err
		}

		return string(templateHtml), nil

	}

	return htmlData, nil
}
