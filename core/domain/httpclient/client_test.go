package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetPage(t *testing.T) {
	// Setup a test server that can serve different content based on the path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var htmlBody string
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			htmlBody = `
				<html>
				<head>
					<title>Home Page</title>
					<link rel="stylesheet" href="/assets/style.css">
				</head>
				<body>
					<h1>Welcome</h1>
					<a href="/about.html">About</a>
					<a href="contact.html">Contact (relative)</a>
					<img src="images/logo.png" />
				</body>
				</html>`
		case "/sub/page.html":
			w.Header().Set("Content-Type", "text/html")
			htmlBody = `
				<html>
				<body>
					<a href="/index.html">Home</a>
					<a href="other.html">Other (relative)</a>
					<a href="../images/banner.png">Banner (parent relative)</a>
					<a href="http://external.com/page">External Site</a>
					<a href="mailto:user@example.com">Mail To</a>
					<script src="/app.js"></script>
				</body>
				</html>`
		case "/no-links":
			w.Header().Set("Content-Type", "text/html")
			htmlBody = `<html><body><p>Nothing to see here.</p></body></html>`
		case "/not-found":
			http.NotFound(w, r)
			return
		default:
			// For any other path, just serve some plain text.
			w.Header().Set("Content-Type", "text/plain")
			htmlBody = "This is not HTML."
		}
		fmt.Fprint(w, htmlBody)
	}))
	defer server.Close()

	client := New(server.Client())

	testCases := []struct {
		name          string
		path          string
		expectedCode  int
		expectedLinks []string // expected links are relative to the server root
	}{
		{
			name:         "extracts links from root page",
			path:         "/",
			expectedCode: http.StatusOK,
			expectedLinks: []string{
				"/assets/style.css",
				"/about.html",
				"/contact.html", // resolves to /contact.html from /
				"/images/logo.png",
			},
		},
		{
			name:         "extracts and resolves links from sub page",
			path:         "/sub/page.html",
			expectedCode: http.StatusOK,
			expectedLinks: []string{
				"/index.html",
				"/sub/other.html", // resolves to /sub/other.html from /sub/page.html
				"/images/banner.png",
				"/app.js",
			},
		},
		{
			name:          "handles page with no links",
			path:          "/no-links",
			expectedCode:  http.StatusOK,
			expectedLinks: []string{},
		},
		{
			name:          "handles 404 not found",
			path:          "/not-found",
			expectedCode:  http.StatusNotFound,
			expectedLinks: []string{},
		},
		{
			name:          "handles non-html content",
			path:          "/style.css",
			expectedCode:  http.StatusOK,
			expectedLinks: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pageURL, err := url.Parse(server.URL + tc.path)
			require.NoError(t, err)

			page, err := client.GetPage(context.Background(), pageURL)
			require.NoError(t, err, "GetPage should not return an error for a valid page")

			// Check status code
			assert.Equal(t, tc.expectedCode, page.Code)

			// Check body
			bodyBytes, err := io.ReadAll(page.Body)
			require.NoError(t, err)
			assert.True(t, len(bodyBytes) > 0, "Page body should not be empty")

			// Check links
			var foundLinks []string
			for _, link := range page.Links {
				// We only care about the path part for comparison
				foundLinks = append(foundLinks, link.RequestURI())
			}

			// Sort for stable comparison
			sort.Strings(tc.expectedLinks)
			sort.Strings(foundLinks)

			assert.Equal(t, tc.expectedLinks, foundLinks)
		})
	}

	t.Run("handles server error", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		errorURL, _ := url.Parse(errorServer.URL)
		errorServer.Close() // Close immediately to simulate connection error

		_, err := client.GetPage(context.Background(), errorURL)
		require.Error(t, err, "GetPage should return an error if the server is unreachable")
	})

	t.Run("handles malformed html", func(t *testing.T) {
		malformedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `<html><a href="/good-link>... oops, unclosed quote`)
		}))
		defer malformedServer.Close()

		pageURL, _ := url.Parse(malformedServer.URL)
		page, err := client.GetPage(context.Background(), pageURL)
		require.NoError(t, err)
		// The Go HTML tokenizer is very robust and might recover, but in this case, the attribute is malformed.
		assert.Empty(t, page.Links, "Should not find links in malformed attributes")
	})
}
