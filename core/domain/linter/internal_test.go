package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinter_BrokenInternalURL_404(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 404,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.NotNil(t, found, "expected broken-internal-url lint for 404")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "404")
}

func TestLinter_BrokenInternalURL_500(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Server Error</title></head>
<body><h1>500</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 500,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.NotNil(t, found, "expected broken-internal-url lint for 500")
	assert.Contains(t, found.Evidence, "500")
}

func TestLinter_BrokenInternalURL_403(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Forbidden</title></head>
<body><h1>403</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 403,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.NotNil(t, found, "expected broken-internal-url lint for 403")
}

func TestLinter_BrokenInternalURL_200OK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 200,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.Nil(t, found, "should not trigger for 200 OK")
}

func TestLinter_BrokenInternalURL_301Redirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 301,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.Nil(t, found, "should not trigger for 3XX redirects")
}

// Site rule tests for crawl errors

func TestSiteLint_BrokenInternalURL_CrawlError(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/",
			StatusCode: 200,
		},
		{
			URL:        "http://example.com/error-page",
			CrawlError: "connection refused",
		},
		{
			URL:        "http://example.com/timeout-page",
			CrawlError: "request timeout exceeded",
			IsTimeout:  true,
		},
	}

	results := linter.RunSiteRules(pages)

	// / is OK - should not trigger
	assert.Nil(t, findSiteLint(results["http://example.com/"], "broken-internal-url-crawl-error"))

	// /error-page had a crawl error - should trigger
	lint := findSiteLint(results["http://example.com/error-page"], "broken-internal-url-crawl-error")
	assert.NotNil(t, lint, "expected broken-internal-url-crawl-error lint for crawl error")
	assert.Equal(t, linter.High, lint.Severity)
	assert.Equal(t, linter.Internal, lint.Category)
	assert.Contains(t, lint.Evidence, "Crawl error")

	// /timeout-page timed out - should trigger with timeout evidence
	lint = findSiteLint(results["http://example.com/timeout-page"], "broken-internal-url-crawl-error")
	assert.NotNil(t, lint, "expected broken-internal-url-crawl-error lint for timeout")
	assert.Contains(t, lint.Evidence, "timed out")
}

func TestSiteLint_BrokenInternalURL_NoCrawlError(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:        "http://example.com/",
			StatusCode: 200,
		},
		{
			URL:        "http://example.com/page",
			StatusCode: 404, // Has status code, so was crawled successfully
		},
	}

	results := linter.RunSiteRules(pages)

	// Neither should trigger the crawl error rule (they were crawled successfully)
	assert.Nil(t, findSiteLint(results["http://example.com/"], "broken-internal-url-crawl-error"))
	assert.Nil(t, findSiteLint(results["http://example.com/page"], "broken-internal-url-crawl-error"))
}
