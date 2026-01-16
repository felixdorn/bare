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
