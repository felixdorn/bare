package reporter

import (
	"embed"
	"html/template"
	"io"
	"time"

	"github.com/felixdorn/bare/core/domain/analyzer"
)

//go:embed templates/*.html
var templateFS embed.FS

// PageReport contains all the data for a single page in the report.
type PageReport struct {
	URL         string
	Title       string
	Description string
	Canonical   string
	StatusCode  int
	Images      []analyzer.Image
}

// Report contains all the data for the full report.
type Report struct {
	SiteURL     string
	GeneratedAt time.Time
	Pages       []PageReport
	TotalPages  int
	TotalImages int
}

// Reporter generates HTML reports from crawl results.
type Reporter struct {
	tmpl *template.Template
}

// New creates a new Reporter instance.
func New() (*Reporter, error) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Reporter{tmpl: tmpl}, nil
}

// Generate writes an HTML report to the given writer.
func (r *Reporter) Generate(w io.Writer, report *Report) error {
	// Calculate totals
	report.TotalPages = len(report.Pages)
	report.TotalImages = 0
	for _, p := range report.Pages {
		report.TotalImages += len(p.Images)
	}

	return r.tmpl.ExecuteTemplate(w, "report.html", report)
}
