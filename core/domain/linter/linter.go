package linter

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/analyzer"
	"github.com/felixdorn/bare/core/domain/url"
)

// NewContext creates a new linting context from raw page data
func NewContext(body []byte, pageURL *url.URL, analysis *analyzer.Analysis) (*Context, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return &Context{
		Doc:      doc,
		URL:      pageURL,
		Body:     body,
		Analysis: analysis,
	}, nil
}

// Check creates a context and runs all rules, returning all lints found
func Check(body []byte, pageURL *url.URL, analysis *analyzer.Analysis) ([]Lint, error) {
	ctx, err := NewContext(body, pageURL, analysis)
	if err != nil {
		return nil, err
	}
	return Run(ctx), nil
}
