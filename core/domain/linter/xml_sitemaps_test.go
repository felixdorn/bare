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

// Tests for noindex detection

func TestIsNoindexHTML_MetaRobots(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Test</title>
<meta name="robots" content="noindex">
</head>
<body><h1>Hello</h1></body>
</html>`)

	assert.True(t, linter.IsNoindexHTML(html))
}

func TestIsNoindexHTML_MetaRobotsNofollow(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Test</title>
<meta name="robots" content="noindex,nofollow">
</head>
<body><h1>Hello</h1></body>
</html>`)

	assert.True(t, linter.IsNoindexHTML(html))
}

func TestIsNoindexHTML_MetaGooglebot(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Test</title>
<meta name="googlebot" content="noindex">
</head>
<body><h1>Hello</h1></body>
</html>`)

	assert.True(t, linter.IsNoindexHTML(html))
}

func TestIsNoindexHTML_NoNoindex(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Test</title>
<meta name="robots" content="index,follow">
</head>
<body><h1>Hello</h1></body>
</html>`)

	assert.False(t, linter.IsNoindexHTML(html))
}

func TestIsNoindexHTML_NoMetaTag(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Test</title>
</head>
<body><h1>Hello</h1></body>
</html>`)

	assert.False(t, linter.IsNoindexHTML(html))
}

// Tests for sitemap-has-noindex-url rule

func TestSiteLint_SitemapHasNoindexURL(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/",
			StatusCode: 200,
			InSitemap:  true,
			IsNoindex:  false,
		},
		{
			URL:        "http://example.com/noindex-page",
			StatusCode: 200,
			InSitemap:  true,
			IsNoindex:  true,
		},
		{
			URL:        "http://example.com/noindex-not-in-sitemap",
			StatusCode: 200,
			InSitemap:  false,
			IsNoindex:  true,
		},
	}

	results := linter.RunSiteRules(pages)

	// /noindex-page is noindex AND in sitemap - should trigger
	lint := findSiteLint(results["http://example.com/noindex-page"], "sitemap-has-noindex-url")
	assert.NotNil(t, lint, "expected sitemap-has-noindex-url lint")
	assert.Equal(t, linter.Critical, lint.Severity)
	assert.Equal(t, linter.XMLSitemaps, lint.Category)

	// / is indexable in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/"], "sitemap-has-noindex-url"))

	// /noindex-not-in-sitemap is noindex but NOT in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/noindex-not-in-sitemap"], "sitemap-has-noindex-url"))
}

// Tests for sitemap-has-4xx-url rule

func TestSiteLint_SitemapHas4xxURL(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/",
			StatusCode: 200,
			InSitemap:  true,
		},
		{
			URL:        "http://example.com/not-found",
			StatusCode: 404,
			InSitemap:  true,
		},
		{
			URL:        "http://example.com/not-in-sitemap",
			StatusCode: 404,
			InSitemap:  false,
		},
	}

	results := linter.RunSiteRules(pages)

	// /not-found is 404 AND in sitemap - should trigger
	lint := findSiteLint(results["http://example.com/not-found"], "sitemap-has-4xx-url")
	assert.NotNil(t, lint, "expected sitemap-has-4xx-url lint")
	assert.Equal(t, linter.Critical, lint.Severity)
	assert.Equal(t, linter.XMLSitemaps, lint.Category)
	assert.Contains(t, lint.Evidence, "404")

	// / is 200 in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/"], "sitemap-has-4xx-url"))

	// /not-in-sitemap is 404 but NOT in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/not-in-sitemap"], "sitemap-has-4xx-url"))
}

func TestSiteLint_SitemapHas4xxURL_Various4xxCodes(t *testing.T) {
	codes := []int{400, 401, 403, 404, 410, 451}

	for _, code := range codes {
		pages := []linter.SiteLintInput{
			{
				URL:        "http://example.com/error",
				StatusCode: code,
				InSitemap:  true,
			},
		}

		results := linter.RunSiteRules(pages)
		lint := findSiteLint(results["http://example.com/error"], "sitemap-has-4xx-url")
		assert.NotNil(t, lint, "expected lint for status code %d", code)
	}
}

// Tests for sitemap-has-canonicalized-url rule

func TestSiteLint_SitemapHasCanonicalizedURL(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/",
			StatusCode: 200,
			InSitemap:  true,
			Canonical:  "http://example.com/", // self-referencing
		},
		{
			URL:        "http://example.com/page-a",
			StatusCode: 200,
			InSitemap:  true,
			Canonical:  "http://example.com/page-b", // points elsewhere
		},
		{
			URL:        "http://example.com/not-in-sitemap",
			StatusCode: 200,
			InSitemap:  false,
			Canonical:  "http://example.com/other", // not in sitemap
		},
		{
			URL:        "http://example.com/no-canonical",
			StatusCode: 200,
			InSitemap:  true,
			Canonical:  "", // no canonical
		},
	}

	results := linter.RunSiteRules(pages)

	// /page-a has non-self canonical AND in sitemap - should trigger
	lint := findSiteLint(results["http://example.com/page-a"], "sitemap-has-canonicalized-url")
	assert.NotNil(t, lint, "expected sitemap-has-canonicalized-url lint")
	assert.Equal(t, linter.High, lint.Severity)
	assert.Equal(t, linter.XMLSitemaps, lint.Category)
	assert.Contains(t, lint.Evidence, "http://example.com/page-b")

	// / has self-referencing canonical - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/"], "sitemap-has-canonicalized-url"))

	// /not-in-sitemap has canonical but NOT in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/not-in-sitemap"], "sitemap-has-canonicalized-url"))

	// /no-canonical has no canonical - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/no-canonical"], "sitemap-has-canonicalized-url"))
}

func TestSiteLint_SitemapHasCanonicalizedURL_ExternalCanonical(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/page",
			StatusCode: 200,
			InSitemap:  true,
			Canonical:  "http://other-site.com/page", // external canonical
		},
	}

	results := linter.RunSiteRules(pages)

	lint := findSiteLint(results["http://example.com/page"], "sitemap-has-canonicalized-url")
	assert.NotNil(t, lint, "expected lint for external canonical")
	assert.Contains(t, lint.Evidence, "http://other-site.com/page")
}

// Tests for sitemap-has-disallowed-url rule

func TestSiteLint_SitemapHasDisallowedURL(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:          "http://example.com/",
			StatusCode:   200,
			InSitemap:    true,
			IsDisallowed: false,
		},
		{
			URL:          "http://example.com/admin",
			StatusCode:   200,
			InSitemap:    true,
			IsDisallowed: true,
		},
		{
			URL:          "http://example.com/private",
			StatusCode:   200,
			InSitemap:    false,
			IsDisallowed: true,
		},
	}

	results := linter.RunSiteRules(pages)

	// /admin is disallowed AND in sitemap - should trigger
	lint := findSiteLint(results["http://example.com/admin"], "sitemap-has-disallowed-url")
	assert.NotNil(t, lint, "expected sitemap-has-disallowed-url lint")
	assert.Equal(t, linter.High, lint.Severity)
	assert.Equal(t, linter.XMLSitemaps, lint.Category)

	// / is allowed in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/"], "sitemap-has-disallowed-url"))

	// /private is disallowed but NOT in sitemap - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/private"], "sitemap-has-disallowed-url"))
}
