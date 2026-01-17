package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/crawler"
	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinter_RedirectsToSelf(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	// Redirect loop: page-a -> page-b -> page-a
	pageURL, _ := url.Parse("http://example.com/page-a")
	chain := []crawler.Redirect{
		{URL: "http://example.com/page-a", StatusCode: 301},
		{URL: "http://example.com/page-b", StatusCode: 302},
		{URL: "http://example.com/page-a", StatusCode: 301}, // Loop back to page-a
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.NotNil(t, found, "expected redirects-to-self lint for redirect loop")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Redirects, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "http://example.com/page-a")
}

func TestLinter_RedirectsToSelf_InChain(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	// Loop in middle of chain: page-b appears twice
	pageURL, _ := url.Parse("http://example.com/page-a")
	chain := []crawler.Redirect{
		{URL: "http://example.com/page-b", StatusCode: 301},
		{URL: "http://example.com/page-c", StatusCode: 302},
		{URL: "http://example.com/page-b", StatusCode: 301}, // Loop back to page-b
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.NotNil(t, found, "expected redirects-to-self lint when loop detected in chain")
}

func TestLinter_RedirectsToSelf_NoSelfRedirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page-a")
	chain := []crawler.Redirect{
		{URL: "http://example.com/page-b", StatusCode: 301},
		{URL: "http://example.com/page-c", StatusCode: 302},
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.Nil(t, found, "should not trigger when no self-redirect exists")
}

func TestLinter_RedirectsToSelf_NoRedirects(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page-a")
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: nil,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.Nil(t, found, "should not trigger when no redirects exist")
}

func TestLinter_RedirectsToSelf_TrailingSlashRedirect(t *testing.T) {
	// Common scenario: /page redirects to /page/ (trailing slash)
	// This should NOT trigger the redirects-to-self lint
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/players/the-anchor")
	chain := []crawler.Redirect{
		{URL: "http://example.com/players/the-anchor", StatusCode: 301}, // Original URL that triggered redirect
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.Nil(t, found, "simple trailing slash redirect should not trigger lint")
}

func TestLinter_RedirectBroken_4XX(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old-page", StatusCode: 301},
	}
	opts := linter.CheckOptions{
		StatusCode:    404,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.NotNil(t, found, "expected redirect-broken lint for 4XX after redirect")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Redirects, found.Category)
	assert.Contains(t, found.Evidence, "404")
}

func TestLinter_RedirectBroken_5XX(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body><h1>500</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old-page", StatusCode: 302},
	}
	opts := linter.CheckOptions{
		StatusCode:    500,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.NotNil(t, found, "expected redirect-broken lint for 5XX after redirect")
	assert.Contains(t, found.Evidence, "500")
}

func TestLinter_RedirectBroken_NoRedirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode:    404,
		RedirectChain: nil, // No redirects
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.Nil(t, found, "should not trigger when there are no redirects")
}

func TestLinter_RedirectBroken_SuccessfulRedirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old-page", StatusCode: 301},
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.Nil(t, found, "should not trigger for successful redirect")
}
