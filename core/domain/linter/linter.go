package linter

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/analyzer"
	"github.com/felixdorn/bare/core/domain/crawler"
	"github.com/felixdorn/bare/core/domain/url"
)

// CheckOptions holds optional parameters for linting
type CheckOptions struct {
	StatusCode    int
	RedirectChain []crawler.Redirect
}

// NewContext creates a new linting context from raw page data
func NewContext(body []byte, pageURL *url.URL, analysis *analyzer.Analysis, opts CheckOptions) (*Context, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return &Context{
		Doc:           doc,
		URL:           pageURL,
		Body:          body,
		Analysis:      analysis,
		StatusCode:    opts.StatusCode,
		RedirectChain: opts.RedirectChain,
	}, nil
}

// Check creates a context and runs all rules, returning all lints found
func Check(body []byte, pageURL *url.URL, analysis *analyzer.Analysis, opts CheckOptions) ([]Lint, error) {
	ctx, err := NewContext(body, pageURL, analysis, opts)
	if err != nil {
		return nil, err
	}
	return Run(ctx), nil
}
