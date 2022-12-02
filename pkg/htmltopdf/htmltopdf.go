package htmltopdf

import (
	"html/template"
	"os"

	"chemin-du-local.bzh/graphql/internal/config"
)

type InvoiceData struct {
	Day             string
	Month           string
	Year            string
	InvoiceNumber   string
	StoreName       string
	StoreSIREN      string
	StoreAddress1   string
	StoreAddress2   string
	StoreTVA        string
	StoreEmail      string
	ServicesContent template.HTML
	BilledDate      string
	Card            string
	TotalHT         string
	TotalTVA        string
	TotalTTC        string
}

func convertToPDF(
	templatePath string,
	outputPath string,
	templateData interface{},
) error {
	r := NewRequestPdf("")

	if err := r.ParseTemplate(templatePath, templateData); err == nil {
		r.GeneratePDF(outputPath)
		return nil
	} else {
		return err
	}
}

func InvoiceToPDF(commerceID string, data InvoiceData) {
	templatePath := config.Cfg.Paths.Static + "/templates/invoice.html"
	folderPath := config.Cfg.Paths.Static + "/commerces/" + commerceID + "/bills"
	os.MkdirAll(folderPath, os.ModePerm)
	outputPath := folderPath + "/" + data.Year + data.Month + data.Day + ".pdf"

	convertToPDF(
		templatePath,
		outputPath,
		data,
	)
}
