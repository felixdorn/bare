package linter

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/analyzer"
	"github.com/felixdorn/bare/core/domain/url"
)

// Context is passed to every rule check and contains everything a rule might need
type Context struct {
	Doc      *goquery.Document
	URL      *url.URL
	Body     []byte
	Analysis *analyzer.Analysis
}

// Lint is a single issue found by a rule
type Lint struct {
	Rule     string
	Message  string
	Severity Severity
	Category Category
	Tag      Tag
	Evidence string
}

// Rule defines a single linting rule
type Rule struct {
	ID       string
	Name     string
	Severity Severity
	Category Category
	Tag      Tag
	Check    func(ctx *Context) []Lint
}

// Emit creates a lint from this rule with optional evidence
func (r *Rule) Emit(evidence string) Lint {
	return Lint{
		Rule:     r.ID,
		Message:  r.Name,
		Severity: r.Severity,
		Category: r.Category,
		Tag:      r.Tag,
		Evidence: evidence,
	}
}
