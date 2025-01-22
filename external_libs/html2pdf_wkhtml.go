package externallibs

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type Generator[T any] struct {
	OutputPath      string             // directory path for generated data
	FinalPdf        string             // the merged pdf name (make sure to include .pdf in the name)
	Template        *template.Template // html template
	Data            []T                // valid data for feeding the template
	HtmlFiles       []*os.File         // list of generated html files
	PdfFiles        []string           // list of generated pdf files
	SingleHtmlFile  bool               // If you want the template to be single file only
	JavascriptToRun string             // javascript you can run before converting html to pdf
	DocumentTitle   string             // the document title after after AI wirtes the <title> tag
}

// Generate pdf file from multiple html templates
func (g *Generator[T]) WkCreatePdf(username string) error {
	// Check if wkhtmltopdf is installed
	_, err := exec.LookPath("wkhtmltopdf")
	if err != nil {
		return fmt.Errorf("wkhtmltopdf is not installed: %v", err)
	}

	err = g.WkGenerateTemplates()
	if err != nil {
		return err
	}

	for _, file := range g.HtmlFiles {
		defer file.Close()
		pdfFilePath := fmt.Sprintf("./%s/%s.pdf", g.OutputPath, g.OutputPath)

		// Get absolute path of the HTML file
		filePath, err := filepath.Abs(file.Name())
		if err != nil {
			return fmt.Errorf("error getting absolute file path: %v", err)
		}

		err = getHtmlFileAndInjectJS(filePath, g.JavascriptToRun)
		if err != nil {
			return err
		}

		// Create wkhtmltopdf command with options
		cmd := exec.Command("wkhtmltopdf",
			"--enable-javascript",
			"--enable-local-file-access", // Allow access to local files
			"--quiet",                    // Reduce output
			"--page-size", "A4",          // Set page size
			"--margin-top", "10mm", // Set margins
			"--margin-bottom", "10mm",
			"--margin-left", "10mm",
			"--margin-right", "10mm",
			"--encoding", "UTF-8", // Ensure proper encoding
			filePath,    // Input HTML file
			pdfFilePath, // Output PDF file
		)

		// Execute the command
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("error generating PDF: %v, output: %s", err, output)
		}

		g.PdfFiles = append(g.PdfFiles, pdfFilePath)
		fmt.Printf("Generated PDF: %s from %s\n", pdfFilePath, filePath)
	}

	// Merge the PDF files
	err = g.WkMergePDFs(g.PdfFiles, g.FinalPdf, username)
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err)
	}

	g.FinalPdf = fmt.Sprintf("pdfs/%s/%s_final.pdf", username, g.OutputPath)
	return nil
}

// Generate html templates from the given data and save them into .g.OutputPath
func (g *Generator[T]) WkGenerateTemplates() error {
	if g.SingleHtmlFile {
		file, err := g.WkCreateHtmlFile(1)
		if err != nil {
			return err
		}
		err = g.Template.Execute(file, g.Data)
		if err != nil {
			return fmt.Errorf("error generating templates: %v", err)
		}
		g.HtmlFiles = append(g.HtmlFiles, file)
		return nil
	}

	// for multiple html files
	for i, v := range g.Data {
		file, err := g.WkCreateHtmlFile(i)
		if err != nil {
			return err
		}
		err = g.Template.Execute(file, v)
		if err != nil {
			return fmt.Errorf("error generating templates: %v", err)
		}
		g.HtmlFiles = append(g.HtmlFiles, file)
	}

	return nil
}

func (g *Generator[T]) WkCreateHtmlFile(id int) (*os.File, error) {
	err := os.MkdirAll(g.OutputPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating output directory: %v", err)
	}

	// name := fmt.Sprintf("./%s/output%d.html", g.OutputPath, id)
	name := fmt.Sprintf("%s/%s.html", g.OutputPath, g.OutputPath)
	file, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("error creating html files: %v", err)
	}

	return file, nil
}

// Delete html and pdf files except the merged pdf
func (g *Generator[T]) WkDeleteFiles() error {
	err := os.RemoveAll(g.OutputPath)
	if err != nil {
		return fmt.Errorf("error deleting files directory: %v", err)
	}
	return nil
}

// Merging all generated pdf together and create the output file
func (g *Generator[T]) WkMergePDFs(inputFiles []string, outputFile string, username string) error {
	dirPath := fmt.Sprintf("pdfs/%s", username)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	err = api.MergeCreateFile(inputFiles, fmt.Sprintf("%s/%s_final.pdf", dirPath, outputFile), false, api.LoadConfiguration())
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err)
	}
	return nil
}

func getHtmlFileAndInjectJS(name string, js string) error {
	// Read the file contents
	content, err := os.ReadFile(name)

	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return fmt.Errorf("error reading the html file")
	}

	// Convert the byte slice to a string
	fileContent := string(content)

	modifiedHTML := injectSelfExecutingJS(fileContent, js)

	err = overrideHTMLFile(name, modifiedHTML)
	if err != nil {
		return fmt.Errorf("error overriding html file: %v", err)
	}

	return nil

}

func injectSelfExecutingJS(htmlStr string, js string) string {
	// Create the self-executing JavaScript script tag
	script := fmt.Sprintf(`
		<script>
			%s
		</script>
	`, js)

	// Find the <head> tag and inject the script just before it closes
	headCloseTag := "</head>"
	if strings.Contains(htmlStr, headCloseTag) {
		return strings.Replace(htmlStr, headCloseTag, script+headCloseTag, 1)
	}

	// If <head> tag is missing, create one and inject the script
	return strings.Replace(htmlStr, "<html>", "<html><head>"+script+"</head>", 1)
}

func overrideHTMLFile(filePath string, newContent string) error {
	// Open the file in write mode and truncate its content
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Write the new content to the file
	_, err = file.WriteString(newContent)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Println("File content overridden successfully!")
	return nil
}
