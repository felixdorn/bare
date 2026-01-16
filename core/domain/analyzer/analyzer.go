package analyzer

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/url"
)

// Analysis contains the results of analyzing a page.
type Analysis struct {
	Title       string
	Description string
	Canonical   string
}

// Analyze parses an HTML body and extracts metadata.
func Analyze(body []byte, pageURL *url.URL) (*Analysis, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	analysis := &Analysis{
		Title:       strings.TrimSpace(doc.Find("head title").First().Text()),
		Description: doc.Find(`head meta[name="description"]`).AttrOr("content", ""),
		Canonical:   doc.Find(`head link[rel="canonical"]`).AttrOr("href", ""),
	}

	return analysis, nil
}
