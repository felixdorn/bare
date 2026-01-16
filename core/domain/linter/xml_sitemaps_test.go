package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/stretchr/testify/assert"
)

// Tests for sitemap parser

func TestParseSitemapURLs_StandardSitemap(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/page1</loc>
  </url>
  <url>
    <loc>https://example.com/page2</loc>
  </url>
  <url>
    <loc>https://example.com/page3</loc>
  </url>
</urlset>`)

	urls := linter.ParseSitemapURLs(content)
	assert.Len(t, urls, 3)
	assert.Contains(t, urls, "https://example.com/page1")
	assert.Contains(t, urls, "https://example.com/page2")
	assert.Contains(t, urls, "https://example.com/page3")
}

func TestParseSitemapURLs_SitemapIndex(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap>
    <loc>https://example.com/sitemap1.xml</loc>
  </sitemap>
  <sitemap>
    <loc>https://example.com/sitemap2.xml</loc>
  </sitemap>
</sitemapindex>`)

	urls := linter.ParseSitemapURLs(content)
	assert.Len(t, urls, 2)
	assert.Contains(t, urls, "https://example.com/sitemap1.xml")
	assert.Contains(t, urls, "https://example.com/sitemap2.xml")
}

func TestParseSitemapURLs_EmptySitemap(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
</urlset>`)

	urls := linter.ParseSitemapURLs(content)
	assert.Empty(t, urls)
}

func TestParseSitemapURLs_InvalidXML(t *testing.T) {
	content := []byte(`not xml at all`)

	urls := linter.ParseSitemapURLs(content)
	assert.Empty(t, urls)
}

func TestParseSitemapURLs_HTMLNotSitemap(t *testing.T) {
	content := []byte(`<!DOCTYPE html>
<html>
<head><title>Not a sitemap</title></head>
<body><h1>Hello</h1></body>
</html>`)

	urls := linter.ParseSitemapURLs(content)
	assert.Empty(t, urls)
}

func TestParseSitemapURLs_WhitespaceInLoc(t *testing.T) {
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>  https://example.com/page1  </loc>
  </url>
</urlset>`)

	urls := linter.ParseSitemapURLs(content)
	assert.Len(t, urls, 1)
	assert.Equal(t, "https://example.com/page1", urls[0])
}

func TestIsSitemapURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/sitemap.xml", true},
		{"https://example.com/SITEMAP.XML", true},
		{"https://example.com/sitemap-posts.xml", true},
		{"https://example.com/post-sitemap.xml", true},
		{"https://example.com/page.html", false},
		{"https://example.com/sitemap", false},
		{"https://example.com/", false},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			assert.Equal(t, tc.expected, linter.IsSitemapURL(tc.url))
		})
	}
}

func TestIsSitemapContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"urlset", `<urlset xmlns="...">`, true},
		{"sitemapindex", `<sitemapindex xmlns="...">`, true},
		{"html", `<!DOCTYPE html><html>`, false},
		{"empty", ``, false},
		{"random xml", `<root><child/></root>`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, linter.IsSitemapContent([]byte(tc.content)))
		})
	}
}

// Tests for sitemap site rules

func TestSiteLint_SitemapHas5xxURL(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/",
			StatusCode: 200,
			InSitemap:  true,
		},
		{
			URL:        "http://example.com/error",
			StatusCode: 500,
			InSitemap:  true,
		},
		{
			URL:        "http://example.com/not-in-sitemap",
			StatusCode: 500,
			InSitemap:  false,
		},
	}

	results := linter.RunSiteRules(pages)

	// /error is 500 AND in sitemap - should trigger
	lint := findSiteLint(results["http://example.com/error"], "sitemap-has-5xx-url")
	assert.NotNil(t, lint, "expected sitemap-has-5xx-url lint")
	assert.Equal(t, linter.Critical, lint.Severity)
	assert.Equal(t, linter.XMLSitemaps, lint.Category)
	assert.Contains(t, lint.Evidence, "500")

	// / is 200 in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/"], "sitemap-has-5xx-url"))

	// /not-in-sitemap is 500 but NOT in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/not-in-sitemap"], "sitemap-has-5xx-url"))
}

func TestSiteLint_SitemapHas5xxURL_Various5xxCodes(t *testing.T) {
	codes := []int{500, 501, 502, 503, 504}

	for _, code := range codes {
		pages := []linter.SiteLintInput{
			{
				URL:        "http://example.com/error",
				StatusCode: code,
				InSitemap:  true,
			},
		}

		results := linter.RunSiteRules(pages)
		lint := findSiteLint(results["http://example.com/error"], "sitemap-has-5xx-url")
		assert.NotNil(t, lint, "expected lint for status code %d", code)
	}
}
