package externallibs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Generate pdf file from multiple html templates
func (g *Generator[T]) WkCreatePdf() error {
	// Check if wkhtmltopdf is installed
	_, err := exec.LookPath("wkhtmltopdf")
	if err != nil {
		return fmt.Errorf("wkhtmltopdf is not installed: %v", err)
	}

	err = g.GenerateTemplates()
	if err != nil {
		return err
	}

	for i, file := range g.HtmlFiles {
		defer file.Close()
		pdfFilePath := fmt.Sprintf("./%s/output%d.pdf", g.OutputPath, i)

		// Get absolute path of the HTML file
		filePath, err := filepath.Abs(file.Name())
		if err != nil {
			return fmt.Errorf("error getting absolute file path: %v", err)
		}

		// Create wkhtmltopdf command with options
		cmd := exec.Command("wkhtmltopdf",
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
	err = g.MergePDFs(g.PdfFiles, g.FinalPdf)
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err)
	}

	return nil
}

// Generate html templates from the given data and save them into .g.OutputPath
func (g *Generator[T]) WkGenerateTemplates() error {
	if g.SingleHtmlFile {
		file, err := g.CreateHtmlFile(1)
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
		file, err := g.CreateHtmlFile(i)
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

	name := fmt.Sprintf("./%s/output%d.html", g.OutputPath, id)
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
func (g *Generator[T]) WkMergePDFs(inputFiles []string, outputFile string) error {
	err := api.MergeCreateFile(inputFiles, outputFile, false, api.LoadConfiguration())
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err)
	}
	return nil
}
