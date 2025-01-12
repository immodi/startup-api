package externallibs

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type Generator[T any] struct {
	OutputPath      string             // directory path for generated data
	FinalPdf        string             // the merged pdf name (make sure to include .pdf in the name)
	Template        *template.Template // html template
	Data            []T                // valid data for feeding the template
	browser         *rod.Browser       // rod browser for auto generating pdf from html views
	HtmlFiles       []*os.File         // list of generated html files
	PdfFiles        []string           // list of generated pdf files
	SingleHtmlFile  bool               // If you want the template to be single file only
	JavascriptToRun string             // javascript you can run before converting html to pdf
	DocumentTitle   string             // the document title after after AI wirtes the <title> tag
}

// Generate pdf file from multible html templates
func (g *Generator[T]) CreatePdf() error {
	err := g.GenerateTemplates()
	if err != nil {
		return err
	}

	l := launcher.New().Headless(true).Leakless(true)
	g.browser = rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer g.browser.MustClose()

	for i, file := range g.HtmlFiles {
		defer file.Close()
		pdfFilePath := fmt.Sprintf("./%s/output%d.pdf", g.OutputPath, i)

		cd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory path : %v", err.Error())
		}

		filePath := filepath.Join(cd, file.Name())
		filePath = strings.ReplaceAll(filePath, "\\", "/")

		// Construct the URL using the file path
		url := "file://" + filePath

		err = g.CapturePDF(g.browser, url, pdfFilePath)
		if err != nil {
			return fmt.Errorf("error capturing PDF files: %v", err.Error())
		}
		g.PdfFiles = append(g.PdfFiles, pdfFilePath)
	}

	// Merge the PDF files
	err = g.MergePDFs(g.PdfFiles, g.FinalPdf)
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err.Error())
	}

	return nil
}

// Generate html templates from the given data and save them into .g.OutputPath
func (g *Generator[T]) GenerateTemplates() error {
	if g.SingleHtmlFile {
		file, err := g.CreateHtmlFile(1)
		if err != nil {
			return err
		}
		err = g.Template.Execute(file, g.Data)
		if err != nil {
			return fmt.Errorf("error generating templates: %v", err.Error())
		}
		g.HtmlFiles = append(g.HtmlFiles, file)
		return nil
	}

	// for multible html files
	for i, v := range g.Data {
		file, err := g.CreateHtmlFile(i)
		if err != nil {
			return err
		}
		err = g.Template.Execute(file, v)
		if err != nil {
			return fmt.Errorf("error generating templates: %v", err.Error())
		}
		g.HtmlFiles = append(g.HtmlFiles, file)
	}

	return nil
}

func (g *Generator[T]) CreateHtmlFile(id int) (*os.File, error) {
	os.Mkdir(g.OutputPath, 0755)
	name := fmt.Sprintf("./%s/output%d.html", g.OutputPath, id)
	file, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("error creating html files: %v", err.Error())
	}
	return file, nil
}

// Delete html and pdf files except the merged pdf
func (g *Generator[T]) DeleteFiles() error {
	err := os.RemoveAll(g.OutputPath)
	if err != nil {
		return fmt.Errorf("error deleting files directory: %v", err)
	}
	return nil
}

// Automate opening a prowser then capture the html page as single pdf file
func (g *Generator[T]) CapturePDF(browser *rod.Browser, htmlUrl, outputPath string) error {
	page, err := browser.Page(proto.TargetCreateTarget{URL: htmlUrl})
	if err != nil {
		return fmt.Errorf("error creating browser page: %v", err)
	}

	// Wait for the page to load completely
	if g.JavascriptToRun != "" {
		page.MustWaitLoad().MustEval(g.JavascriptToRun)
		g.DocumentTitle = page.MustWaitLoad().MustEval(`
			() => {
				try {
					const canvas = document.querySelector("#canvas");
					canvas.classList.remove("overflow-y-scroll");
					const style = document.querySelector("style");
					style.textContent =
						"html, body { overflow: none; scrollbar-width: none; -ms-overflow-style: none; } body::-webkit-scrollbar { display: none; } " +
						style.textContent;

					const whitespaces = document.querySelectorAll(".whitespace-pre-wrap");
					whitespaces.forEach((whitespace) =>
						whitespace.classList.remove("whitespace-pre-wrap")
					);
				} catch (error) {}
			}
		`).String()
	}

	// Generate the PDF after ensuring JavaScript changes have been applied
	pdfDataStream, err := page.MustWait(`() => document.readyState === 'complete'`).PDF(&proto.PagePrintToPDF{
		PreferCSSPageSize: true,
	})

	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, pdfDataStream)
	if err != nil {
		return err
	}

	pdfData := buf.Bytes()
	os.WriteFile(outputPath, pdfData, 0644)
	fmt.Println(htmlUrl, ":::::", outputPath)
	return nil
}

// Merging all generated pdf together and create the output file
func (g *Generator[T]) MergePDFs(inputFiles []string, outputFile string) error {
	// Merge the PDF files
	err := api.MergeCreateFile(inputFiles, outputFile, false, api.LoadConfiguration())
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err.Error())
	}
	return nil
}
