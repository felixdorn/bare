package reporter

import (
	"embed"
	"html/template"
	"io"
	"time"

	"github.com/felixdorn/bare/core/domain/analyzer"
	"github.com/felixdorn/bare/core/domain/linter"
)

//go:embed templates/*.html
var templateFS embed.FS

// InternalLink represents a followed internal link from this page.
type InternalLink struct {
	TargetURL string
	IsFollow  bool // true if the link is followed (not nofollow)
}

// PageReport contains all the data for a single page in the report.
type PageReport struct {
	URL           string
	Title         string
	Description   string
	Canonical     string
	StatusCode    int
	Images        []analyzer.Image
	Lints         []linter.Lint
	InternalLinks []InternalLink // Internal links found on this page
}

// Report contains all the data for the full report.
type Report struct {
	SiteURL     string
	GeneratedAt time.Time
	Pages       []PageReport
	TotalPages  int
	TotalImages int
	TotalLints  int
	LintCounts  map[string]int
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
	report.TotalLints = 0
	report.LintCounts = make(map[string]int)

	for _, p := range report.Pages {
		report.TotalImages += len(p.Images)
		report.TotalLints += len(p.Lints)
		for _, l := range p.Lints {
			report.LintCounts[string(l.Severity)]++
		}
	}

	return r.tmpl.ExecuteTemplate(w, "report.html", report)
}
