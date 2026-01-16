package rewriter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriter_Run(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create some "exported" files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(`
		<html>
		<head><link href="http://example.com/style.css"></head>
		<body>
			<a href="http://example.com/about.html">About</a>
			<a href="http://example.com/contact/">Contact</a>
			<a href="http://external.com/page">External</a>
			<a href="http://example.com/missing.html">Missing</a>
		</body>
		</html>
	`), 0644))

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte(`body { color: blue; }`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "about.html"), []byte(`<h1>About</h1>`), 0644))

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "contact"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "contact", "index.html"), []byte(`<h1>Contact</h1>`), 0644))

	// Create sitemap.xml with absolute URLs
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "sitemap.xml"), []byte(`<?xml version="1.0"?>
<urlset>
	<url><loc>http://example.com/</loc></url>
	<url><loc>http://example.com/about.html</loc></url>
	<url><loc>http://example.com/contact/</loc></url>
</urlset>
	`), 0644))

	baseURL, err := url.Parse("http://example.com")
	require.NoError(t, err)

	r := New(tmpDir, baseURL)
	err = r.Run()
	require.NoError(t, err)

	// Check index.html was rewritten
	content, err := os.ReadFile(filepath.Join(tmpDir, "index.html"))
	require.NoError(t, err)
	html := string(content)

	assert.Contains(t, html, `href="/style.css"`, "style.css should be rewritten to relative")
	assert.Contains(t, html, `href="/about.html"`, "about.html should be rewritten to relative")
	assert.Contains(t, html, `href="/contact/"`, "contact/ should be rewritten to relative")
	assert.Contains(t, html, `href="http://external.com/page"`, "external URL should not be changed")
	assert.Contains(t, html, `href="http://example.com/missing.html"`, "missing file URL should not be changed")

	// Check sitemap.xml was rewritten
	content, err = os.ReadFile(filepath.Join(tmpDir, "sitemap.xml"))
	require.NoError(t, err)
	sitemap := string(content)

	assert.Contains(t, sitemap, `<loc>/</loc>`, "root URL should be rewritten")
	assert.Contains(t, sitemap, `<loc>/about.html</loc>`, "about URL should be rewritten")
	assert.Contains(t, sitemap, `<loc>/contact/</loc>`, "contact URL should be rewritten")
}

func TestRewriter_NoRewrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with norewrite URLs
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(`
		<a href="http://example.com/about.html?norewrite">Keep absolute</a>
		<a href="http://example.com/about.html">Make relative</a>
		<a href="http://example.com/about.html?norewrite&foo=bar">Keep with params</a>
		<a href="http://example.com/about.html?foo=bar&norewrite">Keep with params before</a>
	`), 0644))

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "about.html"), []byte(`<h1>About</h1>`), 0644))

	baseURL, err := url.Parse("http://example.com")
	require.NoError(t, err)

	r := New(tmpDir, baseURL)
	err = r.Run()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "index.html"))
	require.NoError(t, err)
	html := string(content)

	assert.Contains(t, html, `href="http://example.com/about.html"`, "norewrite URL should stay absolute but without ?norewrite")
	assert.Contains(t, html, `href="/about.html"`, "normal URL should be rewritten to relative")
	assert.Contains(t, html, `href="http://example.com/about.html?foo=bar"`, "norewrite with other params should keep params")
	assert.NotContains(t, html, "norewrite", "norewrite param should be removed from all URLs")
}
