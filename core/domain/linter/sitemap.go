package linter

import (
	"encoding/xml"
	"strings"
)

// sitemapURLSet represents a standard sitemap with URL entries.
type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

// sitemapIndex represents a sitemap index file.
type sitemapIndex struct {
	XMLName  xml.Name         `xml:"sitemapindex"`
	Sitemaps []sitemapLocator `xml:"sitemap"`
}

type sitemapLocator struct {
	Loc string `xml:"loc"`
}

// ParseSitemapURLs extracts all URLs from sitemap XML content.
// It handles both regular sitemaps (urlset) and sitemap indexes (sitemapindex).
// For sitemap indexes, it returns the sitemap URLs (not the pages they contain).
func ParseSitemapURLs(content []byte) []string {
	var urls []string

	// Try parsing as a regular sitemap (urlset)
	var urlset sitemapURLSet
	if err := xml.Unmarshal(content, &urlset); err == nil && len(urlset.URLs) > 0 {
		for _, u := range urlset.URLs {
			loc := strings.TrimSpace(u.Loc)
			if loc != "" {
				urls = append(urls, loc)
			}
		}
		return urls
	}

	// Try parsing as a sitemap index
	var index sitemapIndex
	if err := xml.Unmarshal(content, &index); err == nil && len(index.Sitemaps) > 0 {
		for _, s := range index.Sitemaps {
			loc := strings.TrimSpace(s.Loc)
			if loc != "" {
				urls = append(urls, loc)
			}
		}
		return urls
	}

	return urls
}

// IsSitemapURL checks if a URL looks like a sitemap URL.
func IsSitemapURL(urlStr string) bool {
	lower := strings.ToLower(urlStr)
	return strings.HasSuffix(lower, "sitemap.xml") ||
		strings.Contains(lower, "sitemap") && strings.HasSuffix(lower, ".xml")
}

// IsSitemapContent checks if content looks like sitemap XML.
func IsSitemapContent(content []byte) bool {
	// Quick check for XML sitemap markers
	s := string(content)
	return strings.Contains(s, "<urlset") || strings.Contains(s, "<sitemapindex")
}
