package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrawler_DeduplicatesFragments(t *testing.T) {
	// Set up a mock server with a page that has multiple links to the same URL with different fragments
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprintln(w, `<html><body>
				<a href="/page">Page</a>
				<a href="/page#section1">Section 1</a>
				<a href="/page#section2">Section 2</a>
				<a href="/page#section3">Section 3</a>
			</body></html>`)
		case "/page":
			fmt.Fprintln(w, `<html><body><h1>The Page</h1></body></html>`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	// Track how many times each URL is visited
	var visitedURLs []string
	var mu sync.Mutex

	c := New(Config{
		BaseURL:     serverURL,
		WorkerCount: 1, // Single worker for deterministic order
		Entrypoints: []string{"/"},
		Logger:      zerolog.Nop(),
		HTTPClient:  server.Client(),
		OnNewLink: func(page *Page, link Link) error {
			// Follow all internal links
			if link.URL.IsInternal(page.URL) {
				return nil
			}
			return fmt.Errorf("external")
		},
		OnPage: func(page *Page) {
			mu.Lock()
			visitedURLs = append(visitedURLs, page.URL.Path)
			mu.Unlock()
		},
	})

	err = c.Run(context.Background())
	require.NoError(t, err)

	// Should only visit "/" and "/page" once each, not "/page" multiple times
	assert.Len(t, visitedURLs, 2, "Should only visit 2 unique pages, not duplicates with fragments")

	// Count occurrences of /page
	pageCount := 0
	for _, u := range visitedURLs {
		if u == "/page" {
			pageCount++
		}
	}
	assert.Equal(t, 1, pageCount, "/page should only be visited once despite multiple fragment links")
}

func TestCrawler_DeduplicatesHTTPAndHTTPS(t *testing.T) {
	// Set up a mock server that serves content with mixed http/https links
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			// Include both http and https versions of the same link
			// In real scenario, server.URL is http://127.0.0.1:port
			// We'll simulate by including explicit http:// and https:// links
			fmt.Fprintf(w, `<html><body>
				<a href="/about">About (relative)</a>
				<a href="http://%s/about">About (http)</a>
				<a href="https://%s/about">About (https)</a>
			</body></html>`, r.Host, r.Host)
		case "/about":
			fmt.Fprintln(w, `<html><body><h1>About</h1></body></html>`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	// Track how many times each path is visited
	var visitedPaths []string
	var mu sync.Mutex

	c := New(Config{
		BaseURL:     serverURL,
		WorkerCount: 1,
		Entrypoints: []string{"/"},
		Logger:      zerolog.Nop(),
		HTTPClient:  server.Client(),
		OnNewLink: func(page *Page, link Link) error {
			// Follow all links that match our host (ignore scheme difference)
			if link.URL.Host == page.URL.Host {
				return nil
			}
			return fmt.Errorf("external")
		},
		OnPage: func(page *Page) {
			mu.Lock()
			visitedPaths = append(visitedPaths, page.URL.Path)
			mu.Unlock()
		},
	})

	err = c.Run(context.Background())
	require.NoError(t, err)

	// Should only visit "/" and "/about" once each
	assert.Len(t, visitedPaths, 2, "Should only visit 2 unique pages")

	// Count occurrences of /about
	aboutCount := 0
	for _, p := range visitedPaths {
		if p == "/about" {
			aboutCount++
		}
	}
	assert.Equal(t, 1, aboutCount, "/about should only be visited once despite http/https variants")
}
