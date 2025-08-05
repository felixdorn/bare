package url

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURL_ToPath(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		root         string
		expectedPath string
	}{
		{
			name:         "root path",
			url:          "http://example.com/",
			root:         "dist",
			expectedPath: filepath.Join("dist", "index.html"),
		},
		{
			name:         "root path without slash",
			url:          "http://example.com",
			root:         "dist",
			expectedPath: filepath.Join("dist", "index.html"),
		},
		{
			name:         "path with trailing slash",
			url:          "http://example.com/about/",
			root:         "dist",
			expectedPath: filepath.Join("dist", "about", "index.html"),
		},
		{
			name:         "path without trailing slash and no extension",
			url:          "http://example.com/contact",
			root:         "dist",
			expectedPath: filepath.Join("dist", "contact", "index.html"),
		},
		{
			name:         "path with html extension",
			url:          "http://example.com/page.html",
			root:         "dist",
			expectedPath: filepath.Join("dist", "page.html"),
		},
		{
			name:         "path with non-html extension",
			url:          "http://example.com/style.css",
			root:         "dist",
			expectedPath: filepath.Join("dist", "style.css"),
		},
		{
			name:         "path in subdirectory",
			url:          "http://example.com/assets/app.js",
			root:         "dist",
			expectedPath: filepath.Join("dist", "assets", "app.js"),
		},
		{
			name:         "URL with query string should be treated as directory",
			url:          "http://example.com/search?q=test",
			root:         "dist",
			expectedPath: filepath.Join("dist", "search", "index.html"),
		},
		{
			name:         "deeper path without extension",
			url:          "http://example.com/some/deep/path",
			root:         "output",
			expectedPath: filepath.Join("output", "some", "deep", "path", "index.html"),
		},
		{
			name:         "deeper path with trailing slash",
			url:          "http://example.com/some/deep/path/",
			root:         "output",
			expectedPath: filepath.Join("output", "some", "deep", "path", "index.html"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := Parse(tc.url)
			require.NoError(t, err)

			actualPath := u.ToPath(tc.root)
			assert.Equal(t, tc.expectedPath, actualPath)
		})
	}
}
