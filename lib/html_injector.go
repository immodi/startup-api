package lib

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/russross/blackfriday/v2"
)

type Tag struct {
	Index     int
	Type      bool //`opening:1 - closing:0`
	TagLength int
}

func WriteResponseHTML(c echo.Context, app *pocketbase.PocketBase, templateName string, htmlData string, styleTag string, insertStyleTag func(string, string) string) error {
	htmlData = insertStyleTag(RemoveTrailingFreeText(htmlData), styleTag)

	localTemplates := make(map[string]string)
	localTemplates["document"] = "document.html"
	localTemplates["report"] = "report.html"
	localTemplates["paragraph"] = "paragraph.html"
	localTemplates["template"] = "template.html"

	var templatePath string
	if _, ok := localTemplates[templateName]; ok {
		templatePath = localTemplates[templateName]
	} else {
		templatePath = localTemplates["template"]
	}

	templateHtml, err := ReadFileData(fmt.Sprintf("templates/%s", templatePath))
	if err != nil {
		return err
	}

	// Create or open a file to write the output
	file, err := os.Create("data.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a new template and parse the template string
	tmpl, err := template.New("page").Parse(templateHtml)
	if err != nil {
		panic(err)
	}

	// Write the template output to the file
	err = tmpl.Execute(file, struct {
		Content template.HTML
	}{
		Content: template.HTML(htmlData),
	})
	if err != nil {
		panic(err)
	}

	return nil
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
