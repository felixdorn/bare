package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinter_LocalhostLink_127001WithPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://127.0.0.1:1234/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for 127.0.0.1 with port")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.Links, found.Category)
	assert.Contains(t, found.Evidence, "127.0.0.1:1234")
}

func TestLinter_LocalhostLink_127001WithoutPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://127.0.0.1/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for 127.0.0.1 without port")
}

func TestLinter_LocalhostLink_LocalhostWithPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://localhost:8080/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for localhost with port")
	assert.Contains(t, found.Evidence, "localhost:8080")
}

func TestLinter_LocalhostLink_LocalhostWithoutPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://localhost/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for localhost without port")
}

func TestLinter_LocalhostLink_CaseInsensitive(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://LOCALHOST/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "should detect LOCALHOST case-insensitively")
}

func TestLinter_LocalhostLink_MultipleLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://localhost/one">One</a>
<a href="http://127.0.0.1/two">Two</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "localhost-link" {
			count++
		}
	}
	assert.Equal(t, 2, count, "should detect both localhost links")
}

func TestLinter_LocalhostLink_NoLocalhostLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/test">Test</a>
<a href="/relative">Relative</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.Nil(t, found, "should not trigger for normal links")
}

func TestLinter_LocalFileLink_WindowsPath(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="C:\Computer\Sitebulb\page.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for Windows path")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.Links, found.Category)
	assert.Contains(t, found.Evidence, "C:\\Computer")
}

func TestLinter_LocalFileLink_WindowsPathForwardSlash(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="D:/Users/test/file.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for Windows path with forward slash")
}

func TestLinter_LocalFileLink_UNCPath(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="\\servername\path\file.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for UNC path")
	assert.Contains(t, found.Evidence, "\\\\servername")
}

func TestLinter_LocalFileLink_FileProtocol(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="file:///C:/Users/test/file.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for file:// protocol")
	assert.Contains(t, found.Evidence, "file:///")
}

func TestLinter_LocalFileLink_FileProtocolCaseInsensitive(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="FILE:///path/to/file">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "should detect FILE:// case-insensitively")
}

func TestLinter_LocalFileLink_NoLocalPaths(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/test">Test</a>
<a href="/relative/path">Relative</a>
<a href="../parent/path">Parent</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.Nil(t, found, "should not trigger for normal links")
}

func TestLinter_WhitespaceHref_LeadingSpace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href=" https://www.example.com">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for leading space")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Links, found.Category)
}

func TestLinter_WhitespaceHref_TrailingSpace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://www.example.com ">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for trailing space")
}

func TestLinter_WhitespaceHref_BothEnds(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="  https://www.example.com  ">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for whitespace on both ends")
}

func TestLinter_WhitespaceHref_Tabs(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="	https://www.example.com">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for tab character")
}

func TestLinter_WhitespaceHref_Newline(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="
https://www.example.com">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for newline character")
}

func TestLinter_WhitespaceHref_NoWhitespace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://www.example.com">Test</a>
<a href="/relative/path">Relative</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.Nil(t, found, "should not trigger for clean hrefs")
}

func TestLinter_WhitespaceHref_MultipleLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href=" /one">One</a>
<a href="/two ">Two</a>
<a href="/clean">Clean</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "whitespace-href" {
			count++
		}
	}
	assert.Equal(t, 2, count, "should detect both links with whitespace")
}

func TestLinter_NoOutgoingLinks_NoLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<p>This page has no links at all.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "expected no-outgoing-links lint")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
}

func TestLinter_NoOutgoingLinks_OnlyFragmentLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#section1">Section 1</a>
<a href="#section2">Section 2</a>
<p id="section1">Content 1</p>
<p id="section2">Content 2</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "fragment-only links should trigger")
}

func TestLinter_NoOutgoingLinks_OnlyJavascriptLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="javascript:void(0);">Click me</a>
<a href="javascript:doSomething()">Do something</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "javascript: links should trigger")
}

func TestLinter_NoOutgoingLinks_OnlyMailtoLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="mailto:test@example.com">Email us</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "mailto: links should trigger")
}

func TestLinter_NoOutgoingLinks_OnlyTelLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="tel:+1234567890">Call us</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "tel: links should trigger")
}

func TestLinter_NoOutgoingLinks_MixedInvalidLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#test">Anchor</a>
<a href="javascript:void(0);">JS</a>
<a href="mailto:test@test.com">Email</a>
<a href="tel:123">Phone</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "mix of invalid links should trigger")
}

func TestLinter_NoOutgoingLinks_HasValidLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="/about">About us</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.Nil(t, found, "relative link should not trigger")
}

func TestLinter_NoOutgoingLinks_HasExternalLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://external.com/page">External</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.Nil(t, found, "external link should not trigger")
}

func TestLinter_NoOutgoingLinks_ValidAmongInvalid(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#test">Anchor</a>
<a href="javascript:void(0);">JS</a>
<a href="/valid-page">Valid</a>
<a href="mailto:test@test.com">Email</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.Nil(t, found, "one valid link among invalid should not trigger")
}

func TestLinter_MalformedHref_SingleSlash(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http:/example.com">Bad link</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.NotNil(t, found, "expected malformed-href lint for single slash")
	assert.Equal(t, linter.High, found.Severity)
	assert.Contains(t, found.Evidence, "http:/example.com")
}

func TestLinter_MalformedHref_ExtraColon(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http:://example.com">Bad link</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.NotNil(t, found, "expected malformed-href lint for extra colon")
}

func TestLinter_MalformedHref_MisspelledHTTP(t *testing.T) {
	testCases := []string{
		"htp://example.com",
		"htpp://example.com",
		"hhtp://example.com",
	}

	for _, href := range testCases {
		html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="` + href + `">Bad link</a>
</body>
</html>`)

		pageURL, _ := url.Parse("http://example.com/")
		lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
		require.NoError(t, err)

		found := findLint(lints, "malformed-href")
		assert.NotNil(t, found, "expected malformed-href lint for %s", href)
	}
}

func TestLinter_MalformedHref_MisspelledHTTPS(t *testing.T) {
	testCases := []string{
		"htps://example.com",
		"httpss://example.com",
		"htpps://example.com",
	}

	for _, href := range testCases {
		html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="` + href + `">Bad link</a>
</body>
</html>`)

		pageURL, _ := url.Parse("http://example.com/")
		lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
		require.NoError(t, err)

		found := findLint(lints, "malformed-href")
		assert.NotNil(t, found, "expected malformed-href lint for %s", href)
	}
}

func TestLinter_MalformedHref_ValidURLs(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com">HTTP</a>
<a href="https://example.com">HTTPS</a>
<a href="/relative/path">Relative</a>
<a href="../parent">Parent</a>
<a href="page.html">Same dir</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.Nil(t, found, "valid URLs should not trigger")
}

func TestLinter_MalformedHref_IgnoresSpecialSchemes(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#anchor">Anchor</a>
<a href="javascript:void(0)">JS</a>
<a href="mailto:test@test.com">Email</a>
<a href="tel:123">Phone</a>
<a href="/valid">Valid</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.Nil(t, found, "special schemes should not trigger")
}

func TestLinter_NonHTTPProtocol_FTP(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ftp://example.com/file.zip">Download</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.NotNil(t, found, "expected non-http-protocol lint for ftp")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
	assert.Contains(t, found.Evidence, "ftp://")
}

func TestLinter_NonHTTPProtocol_SSH(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ssh://user@example.com">SSH</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.NotNil(t, found, "expected non-http-protocol lint for ssh")
}

func TestLinter_NonHTTPProtocol_WebSocket(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ws://example.com/socket">WebSocket</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.NotNil(t, found, "expected non-http-protocol lint for ws")
}

func TestLinter_MalformedHref_InvalidProtocol(t *testing.T) {
	// gopher and news are not valid protocols - should be malformed
	testCases := []string{
		"gopher://example.com/",
		"news:rec.web",
		"irc://irc.example.com/channel",
	}

	for _, href := range testCases {
		html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="` + href + `">Link</a>
<a href="/other">Other</a>
</body>
</html>`)

		pageURL, _ := url.Parse("http://example.com/")
		lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
		require.NoError(t, err)

		found := findLint(lints, "malformed-href")
		assert.NotNil(t, found, "expected malformed-href lint for invalid protocol %s", href)
	}
}

func TestLinter_NonHTTPProtocol_HTTPValid(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/page">HTTP</a>
<a href="https://example.com/page">HTTPS</a>
<a href="/relative">Relative</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.Nil(t, found, "http/https should not trigger")
}

func TestLinter_NonHTTPProtocol_IgnoresSpecialSchemes(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="mailto:test@test.com">Email</a>
<a href="tel:123">Phone</a>
<a href="javascript:void(0)">JS</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.Nil(t, found, "mailto/tel/javascript should not trigger")
}

func TestLinter_NonHTTPProtocol_FTPNotMalformed(t *testing.T) {
	// FTP should trigger non-http-protocol but NOT malformed-href
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ftp://example.com/file.zip">Download</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	malformed := findLint(lints, "malformed-href")
	assert.Nil(t, malformed, "ftp should NOT trigger malformed-href")

	nonHTTP := findLint(lints, "non-http-protocol")
	assert.NotNil(t, nonHTTP, "ftp SHOULD trigger non-http-protocol")
}
