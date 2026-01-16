package linter

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// IsNoindexHTML checks if HTML content contains a noindex directive.
// It checks for <meta name="robots" content="noindex"> variations.
func IsNoindexHTML(body []byte) bool {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return false
	}

	// Check meta robots tag
	noindex := false
	doc.Find(`meta[name="robots"], meta[name="ROBOTS"]`).Each(func(i int, s *goquery.Selection) {
		content, exists := s.Attr("content")
		if exists && strings.Contains(strings.ToLower(content), "noindex") {
			noindex = true
		}
	})

	// Also check googlebot-specific meta tag
	if !noindex {
		doc.Find(`meta[name="googlebot"], meta[name="GOOGLEBOT"]`).Each(func(i int, s *goquery.Selection) {
			content, exists := s.Attr("content")
			if exists && strings.Contains(strings.ToLower(content), "noindex") {
				noindex = true
			}
		})
	}

	return noindex
}
