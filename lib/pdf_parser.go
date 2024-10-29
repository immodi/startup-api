package lib

import (
	"fmt"
	"html/template"
	externallibs "immodi/startup/external_libs"
	"os"
)

type HtmlParserConfig struct {
	HtmlFileName    string
	JavascriptToRun string
}

func ParsePdfFile(config HtmlParserConfig) (string, error) {
	// directory for saving generated data
	tempDir := "files"

	// the final output
	mergedPdf := "data.pdf"

	data, err := ReadHtmlFileData(config.HtmlFileName)
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
	err = g.CreatePdf()
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldnt create the pdf please try again")
	}

	// delete the generated templates and pdf
	err = g.DeleteFiles()
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

func ReadHtmlFileData(htmlFilePath string) (string, error) {
	f, err := os.ReadFile(htmlFilePath)
	if err != nil {
		return "<html><head></head><body>No Data</body></html>", err
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
