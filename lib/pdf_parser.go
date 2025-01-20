package lib

import (
	"fmt"
	"html/template"
	externallibs "immodi/startup/external_libs"
	"immodi/startup/repo"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

type HtmlParserConfig struct {
	TemplateName    string
	JavascriptToRun string
}

func ParsePdfFile(c echo.Context, app *pocketbase.PocketBase, config HtmlParserConfig) (string, error) {
	// directory for saving generated data
	tempDir := "files"

	// the final output
	mergedPdf := "data.pdf"

	data, err := ReadFileData(config.TemplateName)
	if err != nil {
		fmt.Println(err)
	}

	template, err := template.New("name").Parse(data)
	if err != nil {
		fmt.Println(err)
	}

	// create a pdf generator
	g := externallibs.Generator[any]{
		OutputPath:      tempDir,
		FinalPdf:        mergedPdf,
		Template:        template,
		SingleHtmlFile:  true,
		JavascriptToRun: config.JavascriptToRun,
		DocumentTitle:   "",
	}

	// generate pdf
	err = g.WkCreatePdf()
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldnt create the pdf please try again")
	}

	// delete the generated templates and pdf
	err = g.WkDeleteFiles()
	if err != nil {
		return "", fmt.Errorf("couldnt create the pdf please try again")
	}

	if err := os.Mkdir("pdfs", os.ModePerm); err != nil {
		fmt.Println("'pdfs' dir already exists")
	}

	clearPdfsDirectory()

	newFileName := fmt.Sprintf("pdfs/%s.pdf", g.DocumentTitle)
	os.Rename("data.pdf", newFileName)

	return newFileName, nil
}

func ReadHtmlFileDataFromDB(c echo.Context, app *pocketbase.PocketBase, templateName string) (string, error) {
	// f, err := os.ReadFile(htmlFilePath)
	templateHtml, err := repo.GetUserTemplateByName(c, app, templateName)
	if err != nil {
		return "<html><head></head><body>No Data</body></html>", err
	}

	return templateHtml, nil
}

func ReadFileData(fileName string) (string, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(f), nil
}

func clearPdfsDirectory() {
	dir := "pdfs"

	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	// Iterate over files and delete each one
	for _, file := range files {
		if !file.IsDir() { // Skip directories
			err := os.Remove(dir + "/" + file.Name())
			if err != nil {
				fmt.Println("Error deleting file:", file.Name(), err)
			} else {
				fmt.Println("Deleted file:", file.Name())
			}
		}
	}
}
