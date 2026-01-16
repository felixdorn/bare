package analyzer

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/url"
)

// Image represents an image found on a page.
type Image struct {
	URL    *url.URL
	Alt    string
	Src    string // Original src attribute value
	Width  string // Optional width attribute
	Height string // Optional height attribute
}

// Analysis contains the results of analyzing a page.
type Analysis struct {
	Title       string
	Description string
	Canonical   string
	Images      []Image
}

// Analyze parses an HTML body and extracts metadata and images.
func Analyze(body []byte, pageURL *url.URL) (*Analysis, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	analysis := &Analysis{
		Title:       strings.TrimSpace(doc.Find("head title").First().Text()),
		Description: doc.Find(`head meta[name="description"]`).AttrOr("content", ""),
		Canonical:   doc.Find(`head link[rel="canonical"]`).AttrOr("href", ""),
		Images:      []Image{},
	}

	// Extract images
	doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		srcURL, err := url.Parse(src)
		if err != nil {
			return
		}

		// Only include http/https URLs (or relative URLs)
		if srcURL.Scheme != "" && srcURL.Scheme != "http" && srcURL.Scheme != "https" {
			return
		}

		analysis.Images = append(analysis.Images, Image{
			URL:    pageURL.ResolveReference(srcURL),
			Src:    src,
			Alt:    s.AttrOr("alt", ""),
			Width:  s.AttrOr("width", ""),
			Height: s.AttrOr("height", ""),
		})
	})

	// Extract picture > source srcset images
	doc.Find("source[srcset]").Each(func(i int, s *goquery.Selection) {
		srcset, _ := s.Attr("srcset")
		for _, part := range strings.Split(srcset, ",") {
			fields := strings.Fields(strings.TrimSpace(part))
			if len(fields) == 0 {
				continue
			}
			srcURL, err := url.Parse(fields[0])
			if err != nil {
				continue
			}
			analysis.Images = append(analysis.Images, Image{
				URL: pageURL.ResolveReference(srcURL),
				Src: fields[0],
			})
		}
	})

	return analysis, nil
}
