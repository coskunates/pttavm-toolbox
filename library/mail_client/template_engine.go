package mail_client

import (
	"bytes"
	"fmt"
	"html/template"
	"my_toolbox/library/log"
	"path/filepath"
	"strings"
	textTemplate "text/template"
	"time"
)

// TemplateData represents the data structure for email templates
type TemplateData struct {
	Title     string      `json:"title"`
	Header    HeaderData  `json:"header"`
	Summary   SummaryData `json:"summary"`
	Stats     []StatData  `json:"stats,omitempty"`
	StatusBox *StatusData `json:"status_box,omitempty"`
	Details   DetailData  `json:"details"`
	DataTable *TableData  `json:"data_table,omitempty"`
	Result    ResultData  `json:"result"`
	Footer    FooterData  `json:"footer"`
}

type HeaderData struct {
	Icon     string `json:"icon"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type SummaryData struct {
	Title       string     `json:"title"`
	InfoItems   []InfoItem `json:"info_items"`
	Date        string     `json:"date"`
	RecordCount int        `json:"record_count"`
}

type InfoItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type StatData struct {
	Number string `json:"number"`
	Label  string `json:"label"`
}

type StatusData struct {
	Class   string `json:"class"`
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type DetailData struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	InfoBox     *InfoBox `json:"info_box,omitempty"`
}

type InfoBox struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

type TableData struct {
	Title   string     `json:"title"`
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type ResultData struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	FileName string `json:"file_name"`
	FileLink string `json:"file_link"`
}

type FooterData struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// TemplateEngine handles email template processing
type TemplateEngine struct {
	templateDir string
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(templateDir string) *TemplateEngine {
	if templateDir == "" {
		templateDir = "templates/email"
	}

	return &TemplateEngine{
		templateDir: templateDir,
	}
}

// RenderHTML renders HTML template with given data
func (te *TemplateEngine) RenderHTML(templateName string, data TemplateData) (string, error) {
	templatePath := filepath.Join(te.templateDir, templateName+".html")

	// Custom template functions
	funcMap := template.FuncMap{
		"repeat": strings.Repeat,
		"len":    func(s string) int { return len(s) },
	}

	tmpl, err := template.New(templateName + ".html").Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		log.GetLogger().Error("Failed to parse HTML template", err)
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.GetLogger().Error("Failed to execute HTML template", err)
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderText renders text template with given data
func (te *TemplateEngine) RenderText(templateName string, data TemplateData) (string, error) {
	templatePath := filepath.Join(te.templateDir, templateName+".txt")

	// Custom template functions
	funcMap := textTemplate.FuncMap{
		"repeat": strings.Repeat,
		"len":    func(s string) int { return len(s) },
	}

	tmpl, err := textTemplate.New(templateName + ".txt").Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		log.GetLogger().Error("Failed to parse text template", err)
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.GetLogger().Error("Failed to execute text template", err)
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// CreateMailFromTemplate creates a complete mail message from template
func (te *TemplateEngine) CreateMailFromTemplate(recipients []Recipient, fromEmail, fromName, templateName string, data TemplateData, uniqueArgs map[string]interface{}) (MailMessage, error) {
	textContent, err := te.RenderText(templateName, data)
	if err != nil {
		return MailMessage{}, fmt.Errorf("failed to render text: %w", err)
	}

	content := []ContentItem{
		{Type: "text/plain", Value: textContent},
	}

	return MailMessage{
		Subject:    data.Title,
		Tos:        recipients,
		From:       FromAddress{Name: fromName, Email: fromEmail},
		Content:    content,
		UniqueArgs: uniqueArgs,
	}, nil
}

// Helper functions for different job types

// CreateDryRunReportData creates template data for dry run reports
func CreateDryRunReportData(filePath string, recordCount int, date time.Time) TemplateData {
	fileName := filepath.Base(filePath)
	httpLink := fmt.Sprintf("http://192.168.36.240/reports/dry_run_logs/%s/%02d/%s",
		date.Format("2006"), int(date.Month()), fileName)

	return TemplateData{
		Title: "Dry Run Log Analizi Tamamlandi",
		Summary: SummaryData{
			Date:        date.Format("2006-01-02"),
			RecordCount: recordCount,
		},
		Result: ResultData{
			FileName: fileName,
			FileLink: httpLink,
		},
		Footer: FooterData{
			Message:   "Bu rapor otomatik olarak oluşturulmuştur.",
			Signature: fmt.Sprintf("Dry Run Log Analyzer - %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}
}

// getStatusBoxData returns appropriate status box based on record count
func getStatusBoxData(recordCount int) *StatusData {
	if recordCount == 0 {
		return &StatusData{
			Class:   "success",
			Icon:    "✅",
			Title:   "Durum",
			Message: "Hiç komisyon farklılığı tespit edilmedi. Sistem normal çalışıyor.",
		}
	} else if recordCount < 100 {
		return &StatusData{
			Class:   "warning",
			Icon:    "⚠️",
			Title:   "Durum",
			Message: "Az sayıda komisyon farklılığı tespit edildi. İnceleme önerilir.",
		}
	} else {
		return &StatusData{
			Class:   "warning",
			Icon:    "🚨",
			Title:   "Durum",
			Message: "Yüksek sayıda komisyon farklılığı tespit edildi. Acil inceleme gerekli!",
		}
	}
}

// CreateGenericReportData creates template data for generic reports
func CreateGenericReportData(title, description string, stats []StatData, customData map[string]interface{}) TemplateData {
	return TemplateData{
		Title: title,
		Header: HeaderData{
			Icon:     "📊",
			Title:    title,
			Subtitle: "Otomatik Rapor",
		},
		Summary: SummaryData{
			Title: "Rapor Özeti",
			InfoItems: []InfoItem{
				{Label: "📅 Oluşturma Tarihi", Value: time.Now().Format("2006-01-02 15:04:05")},
				{Label: "📋 Açıklama", Value: description},
			},
		},
		Stats: stats,
		Details: DetailData{
			Title:       "📋 Detaylar",
			Description: description,
		},
		Result: ResultData{
			Title:   "📊 Sonuç",
			Message: "Rapor başarıyla oluşturulmuştur.",
		},
		Footer: FooterData{
			Message:   "Bu rapor otomatik olarak oluşturulmuştur.",
			Signature: fmt.Sprintf("System Reporter - %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}
}
