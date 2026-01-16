package analyzer

import (
	"bytes"
	"strings"

	"github.com/felixdorn/bare/core/domain/url"
	"golang.org/x/net/html"
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
	Images []Image
	// Future: lint results, OpenGraph data, etc.
}

// Analyze parses an HTML body and extracts images with their metadata.
func Analyze(body []byte, pageURL *url.URL) (*Analysis, error) {
	analysis := &Analysis{
		Images: []Image{},
	}

	z := html.NewTokenizer(bytes.NewReader(body))

	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			break
		}

		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		t := z.Token()

		if t.Data == "img" {
			img := parseImageTag(t, pageURL)
			if img != nil {
				analysis.Images = append(analysis.Images, *img)
			}
		}

		// Also handle picture > source elements
		if t.Data == "source" {
			// Check if it has srcset (picture element source)
			for _, attr := range t.Attr {
				if attr.Key == "srcset" {
					// Parse srcset - take the first URL
					srcset := strings.TrimSpace(attr.Val)
					if srcset != "" {
						parts := strings.Split(srcset, ",")
						for _, part := range parts {
							part = strings.TrimSpace(part)
							fields := strings.Fields(part)
							if len(fields) > 0 {
								srcURL, err := url.Parse(fields[0])
								if err == nil {
									resolvedURL := pageURL.ResolveReference(srcURL)
									analysis.Images = append(analysis.Images, Image{
										URL: resolvedURL,
										Src: fields[0],
										Alt: "", // source elements don't have alt
									})
								}
							}
						}
					}
					break
				}
			}
		}
	}

	return analysis, nil
}

// parseImageTag extracts image information from an img tag.
func parseImageTag(t html.Token, pageURL *url.URL) *Image {
	var src, alt, width, height string

	for _, attr := range t.Attr {
		switch attr.Key {
		case "src":
			src = attr.Val
		case "alt":
			alt = attr.Val
		case "width":
			width = attr.Val
		case "height":
			height = attr.Val
		}
	}

	if src == "" {
		return nil
	}

	srcURL, err := url.Parse(src)
	if err != nil {
		return nil
	}

	// Only include http/https URLs (or relative URLs)
	if srcURL.Scheme != "" && srcURL.Scheme != "http" && srcURL.Scheme != "https" {
		return nil
	}

	resolvedURL := pageURL.ResolveReference(srcURL)

	return &Image{
		URL:    resolvedURL,
		Src:    src,
		Alt:    alt,
		Width:  width,
		Height: height,
	}
}
