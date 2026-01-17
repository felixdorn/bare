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

// Tracking parameters tests

func TestLinter_TrackingParameters_UTM(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page?utm_medium=email")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.NotNil(t, found, "expected tracking-parameters lint for utm_medium")
	assert.Equal(t, linter.Medium, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "utm_medium")
}

func TestLinter_TrackingParameters_MultipleUTM(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page?utm_source=newsletter&utm_medium=email&utm_campaign=spring")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.NotNil(t, found, "expected tracking-parameters lint")
	assert.Contains(t, found.Evidence, "utm_source")
	assert.Contains(t, found.Evidence, "utm_medium")
	assert.Contains(t, found.Evidence, "utm_campaign")
}

func TestLinter_TrackingParameters_GoogleAds(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page?gclid=abc123")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.NotNil(t, found, "expected tracking-parameters lint for gclid")
	assert.Contains(t, found.Evidence, "gclid")
}

func TestLinter_TrackingParameters_Facebook(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page?fbclid=xyz789")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.NotNil(t, found, "expected tracking-parameters lint for fbclid")
	assert.Contains(t, found.Evidence, "fbclid")
}

func TestLinter_TrackingParameters_Microsoft(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page?msclkid=def456")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.NotNil(t, found, "expected tracking-parameters lint for msclkid")
	assert.Contains(t, found.Evidence, "msclkid")
}

func TestLinter_TrackingParameters_NoTrackingParams(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page?id=123&category=shoes")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.Nil(t, found, "should not trigger for non-tracking query params")
}

func TestLinter_TrackingParameters_NoQueryString(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.Nil(t, found, "should not trigger for URL without query string")
}

func TestLinter_TrackingParameters_MixedParams(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	// Mix of tracking and non-tracking params
	pageURL, _ := url.Parse("http://example.com/page?id=123&utm_source=google&category=shoes")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "tracking-parameters")
	assert.NotNil(t, found, "expected tracking-parameters lint")
	assert.Contains(t, found.Evidence, "utm_source")
	assert.NotContains(t, found.Evidence, "id")
	assert.NotContains(t, found.Evidence, "category")
}

// Non-ASCII URL tests

func TestLinter_NonASCIIURL_SpecialCharsInPath(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/øê.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "non-ascii-url")
	assert.NotNil(t, found, "expected non-ascii-url lint for special characters")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
}

func TestLinter_NonASCIIURL_ChineseChars(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/中央大学.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "non-ascii-url")
	assert.NotNil(t, found, "expected non-ascii-url lint for Chinese characters")
}

func TestLinter_NonASCIIURL_Thorn(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/þorn.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "non-ascii-url")
	assert.NotNil(t, found, "expected non-ascii-url lint for thorn character")
}

func TestLinter_NonASCIIURL_InQueryString(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/page.html?大学")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "non-ascii-url")
	assert.NotNil(t, found, "expected non-ascii-url lint for non-ASCII in query string")
}

func TestLinter_NonASCIIURL_ASCIIOnlyOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/page.html?id=123")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "non-ascii-url")
	assert.Nil(t, found, "should not trigger for ASCII-only URL")
}

// Double slash URL tests

func TestLinter_DoubleSlashURL_AfterDomain(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com//folder/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "double-slash-url")
	assert.NotNil(t, found, "expected double-slash-url lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
}

func TestLinter_DoubleSlashURL_AfterDirectory(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder//page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "double-slash-url")
	assert.NotNil(t, found, "expected double-slash-url lint for double slash after directory")
}

func TestLinter_DoubleSlashURL_TripleSlash(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder///page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "double-slash-url")
	assert.NotNil(t, found, "expected double-slash-url lint for triple slash")
}

func TestLinter_DoubleSlashURL_SingleSlashOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "double-slash-url")
	assert.Nil(t, found, "should not trigger for normal URL with single slashes")
}

func TestLinter_DoubleSlashURL_RootOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "double-slash-url")
	assert.Nil(t, found, "should not trigger for root URL")
}

// Uppercase URL tests

func TestLinter_UppercaseURL_UppercaseFolder(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/Folder/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "uppercase-url")
	assert.NotNil(t, found, "expected uppercase-url lint")
	assert.Equal(t, linter.Medium, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
	assert.Contains(t, found.Evidence, "/Folder/page.html")
}

func TestLinter_UppercaseURL_UppercasePage(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/Page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "uppercase-url")
	assert.NotNil(t, found, "expected uppercase-url lint for uppercase page name")
	assert.Contains(t, found.Evidence, "/folder/Page.html")
}

func TestLinter_UppercaseURL_UppercaseExtension(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/page.HTML")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "uppercase-url")
	assert.NotNil(t, found, "expected uppercase-url lint for uppercase extension")
	assert.Contains(t, found.Evidence, "/folder/page.HTML")
}

func TestLinter_UppercaseURL_LowercaseOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "uppercase-url")
	assert.Nil(t, found, "should not trigger for lowercase URL")
}

func TestLinter_UppercaseURL_RootOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "uppercase-url")
	assert.Nil(t, found, "should not trigger for root URL")
}

func TestLinter_UppercaseURL_IgnoresDomain(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	// Domain has uppercase but path is lowercase - should NOT trigger
	pageURL, _ := url.Parse("http://Example.COM/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "uppercase-url")
	assert.Nil(t, found, "should not trigger for uppercase in domain only")
}

// Whitespace URL tests

func TestLinter_WhitespaceURL_ActualSpace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder path/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.NotNil(t, found, "expected whitespace-url lint for space in path")
	assert.Equal(t, linter.Medium, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
}

func TestLinter_WhitespaceURL_PlusEncoded(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder+path/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.NotNil(t, found, "expected whitespace-url lint for + in path")
	assert.Contains(t, found.Evidence, "+")
}

func TestLinter_WhitespaceURL_PercentEncoded(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder%20path/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.NotNil(t, found, "expected whitespace-url lint for %20 in path")
	// Go's URL parser decodes %20 to space, so evidence will contain decoded path
}

func TestLinter_WhitespaceURL_PercentEncodedUppercase(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	// %20 can also appear as %2F or other cases
	pageURL, _ := url.Parse("http://example.com/folder%20path/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.NotNil(t, found, "expected whitespace-url lint for %20 uppercase")
}

func TestLinter_WhitespaceURL_NoWhitespace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/folder/page.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.Nil(t, found, "should not trigger for URL without whitespace")
}

func TestLinter_WhitespaceURL_HyphenOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	// Hyphens are common word separators and should NOT trigger
	pageURL, _ := url.Parse("http://example.com/folder-path/page-name.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.Nil(t, found, "should not trigger for hyphens")
}

func TestLinter_WhitespaceURL_UnderscoreOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1></body>
</html>`)

	// Underscores should NOT trigger
	pageURL, _ := url.Parse("http://example.com/folder_path/page_name.html")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{StatusCode: 200})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-url")
	assert.Nil(t, found, "should not trigger for underscores")
}
