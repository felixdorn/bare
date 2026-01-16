package exporter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport_Run(t *testing.T) {
	// 1. Set up a mock HTTP server to simulate a website
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprintln(w, `<html><head><link rel="stylesheet" href="/style.css"></head><body><a href="/about.html">About</a><a href="/secret.html">Secret</a></body></html>`)
		case "/about.html":
			fmt.Fprintln(w, `<html><body><h1>About Us</h1><a href="/">Home</a></body></html>`)
		case "/style.css":
			w.Header().Set("Content-Type", "text/css")
			fmt.Fprintln(w, `body { color: blue; }`)
		case "/secret.html":
			fmt.Fprintln(w, `This is a secret page.`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// 2. Create a temporary directory for the output
	outputDir := t.TempDir()

	// 3. Configure the exporter
	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err, "Failed to parse server URL")

	conf := config.NewDefaultConfig()
	conf.URL = serverURL
	conf.Output = outputDir
	conf.Pages.Entrypoints = url.Paths{"/"}
	// Exclude should prevent /secret.html from being crawled or saved
	conf.Pages.Exclude = url.Paths{"/secret.html"}
	// Treat the homepage as extract-only
	conf.Pages.ExtractOnly = url.Paths{"/"}

	// 4. Create the exporter instance with a Nop logger to keep test output clean
	// Use the mock server's client to ensure requests go to our test server
	log := zerolog.Nop()
	export := NewExport(conf, log, server.Client())

	// 6. Run the exporter
	err = export.Run(context.Background())
	require.NoError(t, err, "Exporter run failed")

	// 7. Verify the output directory
	// Check that the expected files were created. /about.html and /style.css
	// should be there because they were linked from the (extract-only) homepage.
	assert.FileExists(t, filepath.Join(outputDir, "about.html"))
	assert.FileExists(t, filepath.Join(outputDir, "style.css"))

	// Check that the homepage itself was *not* created because it's extract-only
	assert.NoFileExists(t, filepath.Join(outputDir, "index.html"))
	// Check that the excluded file was *not* created
	assert.NoFileExists(t, filepath.Join(outputDir, "secret.html"))

	// Optionally, check the contents of a created file to be sure
	aboutContent, err := os.ReadFile(filepath.Join(outputDir, "about.html"))
	require.NoError(t, err)
	assert.Contains(t, string(aboutContent), `<h1>About Us</h1>`, "About.html content is not as expected")
}
