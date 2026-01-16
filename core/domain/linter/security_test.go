package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinter_MixedContent_HTTPImage(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="http://example.com/image.png">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP image on HTTPS page")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.Security, found.Category)
	assert.Contains(t, found.Evidence, "http://example.com/image.png")
}

func TestLinter_MixedContent_HTTPScript(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<script src="http://cdn.example.com/script.js"></script>
</head>
<body>
<h1>Hello</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP script on HTTPS page")
	assert.Contains(t, found.Evidence, "http://cdn.example.com/script.js")
}

func TestLinter_MixedContent_HTTPStylesheet(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<link rel="stylesheet" href="http://cdn.example.com/styles.css">
</head>
<body>
<h1>Hello</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP stylesheet on HTTPS page")
	assert.Contains(t, found.Evidence, "http://cdn.example.com/styles.css")
}

func TestLinter_MixedContent_HTTPVideo(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<video src="http://media.example.com/video.mp4"></video>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP video on HTTPS page")
	assert.Contains(t, found.Evidence, "http://media.example.com/video.mp4")
}

func TestLinter_MixedContent_HTTPAudio(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<audio src="http://media.example.com/audio.mp3"></audio>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP audio on HTTPS page")
	assert.Contains(t, found.Evidence, "http://media.example.com/audio.mp3")
}

func TestLinter_MixedContent_HTTPSource(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<video>
  <source src="http://media.example.com/video.webm" type="video/webm">
</video>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP source on HTTPS page")
	assert.Contains(t, found.Evidence, "http://media.example.com/video.webm")
}

func TestLinter_MixedContent_HTTPIframe(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<iframe src="http://embed.example.com/widget"></iframe>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP iframe on HTTPS page")
	assert.Contains(t, found.Evidence, "http://embed.example.com/widget")
}

func TestLinter_MixedContent_HTTPObject(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<object data="http://example.com/object.swf"></object>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP object on HTTPS page")
	assert.Contains(t, found.Evidence, "http://example.com/object.swf")
}

func TestLinter_MixedContent_HTTPEmbed(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<embed src="http://example.com/embed.swf">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP embed on HTTPS page")
	assert.Contains(t, found.Evidence, "http://example.com/embed.swf")
}

func TestLinter_MixedContent_HTTPVideoPoster(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<video poster="http://example.com/poster.jpg" src="https://example.com/video.mp4"></video>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP video poster on HTTPS page")
	assert.Contains(t, found.Evidence, "http://example.com/poster.jpg")
}

func TestLinter_MixedContent_HTTPSrcset(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img srcset="http://example.com/small.jpg 480w, https://example.com/large.jpg 800w">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.NotNil(t, found, "expected mixed-content lint for HTTP URL in srcset on HTTPS page")
	assert.Contains(t, found.Evidence, "http://example.com/small.jpg")
}

func TestLinter_MixedContent_NotOnHTTPPage(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="http://example.com/image.png">
</body>
</html>`)

	// HTTP page - should NOT trigger mixed content
	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.Nil(t, found, "mixed-content should not trigger on HTTP pages")
}

func TestLinter_MixedContent_HTTPSResourcesOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<script src="https://cdn.example.com/script.js"></script>
<link rel="stylesheet" href="https://cdn.example.com/styles.css">
</head>
<body>
<h1>Hello</h1>
<img src="https://example.com/image.png">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.Nil(t, found, "HTTPS resources should not trigger mixed-content")
}

func TestLinter_MixedContent_RelativeURLsOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<script src="/js/script.js"></script>
<link rel="stylesheet" href="/css/styles.css">
</head>
<body>
<h1>Hello</h1>
<img src="/images/image.png">
<img src="../images/other.png">
<img src="image.png">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.Nil(t, found, "relative URLs should not trigger mixed-content on HTTPS page")
}

func TestLinter_MixedContent_ProtocolRelativeURLsOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<script src="//cdn.example.com/script.js"></script>
</head>
<body>
<h1>Hello</h1>
<img src="//example.com/image.png">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.Nil(t, found, "protocol-relative URLs should inherit HTTPS and not trigger")
}

func TestLinter_MixedContent_DataURLsOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.Nil(t, found, "data: URLs should not trigger mixed-content")
}

func TestLinter_MixedContent_BlobURLsOK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="blob:https://example.com/12345">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "mixed-content")
	assert.Nil(t, found, "blob: URLs should not trigger mixed-content")
}

func TestLinter_MixedContent_MultipleResources(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<script src="http://cdn1.example.com/script.js"></script>
<link rel="stylesheet" href="http://cdn2.example.com/styles.css">
</head>
<body>
<h1>Hello</h1>
<img src="http://images.example.com/image1.png">
<img src="http://images.example.com/image2.png">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "mixed-content" {
			count++
		}
	}
	assert.Equal(t, 4, count, "should detect all 4 HTTP resources")
}

func TestLinter_MixedContent_DuplicateURLsReportedOnce(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="http://example.com/image.png">
<img src="http://example.com/image.png">
<img src="http://example.com/image.png">
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "mixed-content" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate URLs should only be reported once")
}

// Internal HTTP URL tests

func TestLinter_InternalHTTPURL_HTTPWith200(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 200,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "internal-http-url")
	assert.NotNil(t, found, "expected internal-http-url lint for HTTP URL with 200")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.Security, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
}

func TestLinter_InternalHTTPURL_HTTPSWith200(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 200,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "internal-http-url")
	assert.Nil(t, found, "HTTPS URLs should not trigger internal-http-url")
}

func TestLinter_InternalHTTPURL_HTTPWith404(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body>
<h1>404</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 404,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "internal-http-url")
	assert.Nil(t, found, "HTTP URLs with non-200 status should not trigger")
}

func TestLinter_InternalHTTPURL_HTTPWith301(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Redirect</title></head>
<body>
<h1>Moved</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/old-page")
	opts := linter.CheckOptions{
		StatusCode: 301,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "internal-http-url")
	assert.Nil(t, found, "HTTP URLs with redirect status should not trigger")
}

func TestLinter_InternalHTTPURL_HTTPWith500(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body>
<h1>Server Error</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 500,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "internal-http-url")
	assert.Nil(t, found, "HTTP URLs with 500 status should not trigger")
}

// HTTPS links to HTTP tests

func TestLinter_HTTPSLinksToHTTP_InternalHTTPLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/other-page">Other Page</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-links-to-http")
	assert.NotNil(t, found, "expected https-links-to-http lint for internal HTTP link")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Security, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "http://example.com/other-page")
}

func TestLinter_HTTPSLinksToHTTP_ExternalHTTPLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://other-site.com/page">External Page</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-links-to-http")
	assert.Nil(t, found, "external HTTP links should not trigger")
}

func TestLinter_HTTPSLinksToHTTP_InternalHTTPSLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://example.com/other-page">Other Page</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-links-to-http")
	assert.Nil(t, found, "internal HTTPS links should not trigger")
}

func TestLinter_HTTPSLinksToHTTP_RelativeLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="/other-page">Other Page</a>
<a href="../parent">Parent</a>
<a href="sibling.html">Sibling</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-links-to-http")
	assert.Nil(t, found, "relative links should not trigger (inherit HTTPS)")
}

func TestLinter_HTTPSLinksToHTTP_HTTPPage(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/other-page">Other Page</a>
</body>
</html>`)

	// HTTP page - should not trigger this rule
	pageURL, _ := url.Parse("http://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-links-to-http")
	assert.Nil(t, found, "HTTP pages should not trigger this rule")
}

func TestLinter_HTTPSLinksToHTTP_MultipleInternalHTTPLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/page1">Page 1</a>
<a href="http://example.com/page2">Page 2</a>
<a href="https://example.com/page3">Page 3</a>
<a href="http://other-site.com/external">External</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "https-links-to-http" {
			count++
		}
	}
	assert.Equal(t, 2, count, "should detect both internal HTTP links")
}

func TestLinter_HTTPSLinksToHTTP_DuplicateLinksReportedOnce(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/page1">Link 1</a>
<a href="http://example.com/page1">Link 2</a>
<a href="http://example.com/page1">Link 3</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "https-links-to-http" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate links should only be reported once")
}

func TestLinter_HTTPSLinksToHTTP_SkipsSpecialSchemes(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="javascript:void(0)">JS</a>
<a href="mailto:test@example.com">Email</a>
<a href="tel:+1234567890">Phone</a>
<a href="#section">Anchor</a>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-links-to-http")
	assert.Nil(t, found, "special schemes should not trigger")
}

// HTTPS form to HTTP tests

func TestLinter_HTTPSFormToHTTP_FormPOST(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://example.com/form_submit.php" method="POST">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.NotNil(t, found, "expected https-form-to-http lint for form POST to HTTP")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Security, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "http://example.com/form_submit.php")
}

func TestLinter_HTTPSFormToHTTP_FormGET(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://example.com/search" method="GET">
  <input type="text" name="q">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.NotNil(t, found, "expected https-form-to-http lint for form GET to HTTP")
	assert.Contains(t, found.Evidence, "http://example.com/search")
}

func TestLinter_HTTPSFormToHTTP_FormNoMethod(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://example.com/form_submit.php">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.NotNil(t, found, "expected https-form-to-http lint for form without method to HTTP")
}

func TestLinter_HTTPSFormToHTTP_FormToHTTPS(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="https://example.com/form_submit.php" method="POST">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.Nil(t, found, "form to HTTPS should not trigger")
}

func TestLinter_HTTPSFormToHTTP_RelativeAction(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="/submit" method="POST">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.Nil(t, found, "relative action should inherit HTTPS and not trigger")
}

func TestLinter_HTTPSFormToHTTP_HTTPPage(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://example.com/form_submit.php" method="POST">
  <input type="submit">
</form>
</body>
</html>`)

	// HTTP page - should not trigger this rule
	pageURL, _ := url.Parse("http://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.Nil(t, found, "HTTP pages should not trigger this rule")
}

func TestLinter_HTTPSFormToHTTP_MultipleForms(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://example.com/form1" method="POST">
  <input type="submit">
</form>
<form action="http://example.com/form2" method="POST">
  <input type="submit">
</form>
<form action="https://example.com/form3" method="POST">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "https-form-to-http" {
			count++
		}
	}
	assert.Equal(t, 2, count, "should detect both HTTP form actions")
}

func TestLinter_HTTPSFormToHTTP_DuplicateActionsReportedOnce(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://example.com/submit" method="POST">
  <input type="submit">
</form>
<form action="http://example.com/submit" method="GET">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "https-form-to-http" {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate actions should only be reported once")
}

func TestLinter_HTTPSFormToHTTP_ExternalHTTPForm(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="http://other-site.com/submit" method="POST">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	// External HTTP forms are also a security issue
	found := findLint(lints, "https-form-to-http")
	assert.NotNil(t, found, "external HTTP form actions should also trigger")
}

func TestLinter_HTTPSFormToHTTP_JavascriptAction(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<form action="javascript:submitForm()">
  <input type="submit">
</form>
</body>
</html>`)

	pageURL, _ := url.Parse("https://example.com/page")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "https-form-to-http")
	assert.Nil(t, found, "javascript: action should not trigger")
}
