package lib

import (
	"fmt"
	"html/template"
	"os"
)

type Data struct{}

func ParsePdfFile(htmlFileName string) {
	// directory for saving generated data
	tempDir := "files"

	// the final output
	mergedPdf := "data.pdf"

	data, err := ReadHtmlFileData(htmlFileName)
	if err != nil {
		fmt.Println(err)
	}

	template, err := template.New("name").Parse(data)
	if err != nil {
		fmt.Println(err)
	}

	// create a pdf generator
	g := Generator[any]{
		OutputPath:     tempDir,
		FinalPdf:       mergedPdf,
		Template:       template,
		SingleHtmlFile: true,
	}

	// generate pdf
	err = g.CreatePdf()
	if err != nil {
		fmt.Println(err)
	}

	// delete the generated templates and pdf
	err = g.DeleteFiles()
	if err != nil {
		fmt.Println(err)
	}
}

func ReadHtmlFileData(htmlFilePath string) (string, error) {
	f, err := os.ReadFile(htmlFilePath)
	if err != nil {
		return "<html><head></head><body>No Data</body></html>", err
	}

	return string(f), nil
}
