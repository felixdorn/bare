package rewriter

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriter_relativize(t *testing.T) {
	baseURL, err := url.Parse("http://example.com/base/")
	require.NoError(t, err, "Failed to parse base URL")

	r := New("dist", baseURL)

	testCases := []struct {
		name          string
		inputLink     string
		expectedLink  string
		expectChanged bool
	}{
		{
			name:          "absolute internal URL",
			inputLink:     "http://example.com/about.html",
			expectedLink:  "/about.html",
			expectChanged: true,
		},
		{
			name:          "absolute internal URL with subdirectory",
			inputLink:     "http://example.com/blog/my-post",
			expectedLink:  "/blog/my-post",
			expectChanged: true,
		},
		{
			name:          "absolute internal root URL",
			inputLink:     "http://example.com/",
			expectedLink:  "/",
			expectChanged: true,
		},
		{
			name:          "absolute internal URL with query and fragment",
			inputLink:     "http://example.com/search?q=term#results",
			expectedLink:  "/search?q=term#results",
			expectChanged: true,
		},
		{
			name:          "absolute internal https URL",
			inputLink:     "https://example.com/secure",
			expectedLink:  "/secure",
			expectChanged: true,
		},
		{
			name:          "external URL",
			inputLink:     "http://othersite.com/page",
			expectedLink:  "http://othersite.com/page",
			expectChanged: false,
		},
		{
			name:          "relative URL",
			inputLink:     "contact.html",
			expectedLink:  "contact.html",
			expectChanged: false,
		},
		{
			name:          "root-relative URL",
			inputLink:     "/assets/style.css",
			expectedLink:  "/assets/style.css",
			expectChanged: false,
		},
		{
			name:          "URL with different scheme",
			inputLink:     "mailto:user@example.com",
			expectedLink:  "mailto:user@example.com",
			expectChanged: false,
		},
		{
			name:          "malformed URL",
			inputLink:     "http://[::1]:namedport",
			expectedLink:  "http://[::1]:namedport",
			expectChanged: false,
		},
		{
			name:          "empty link",
			inputLink:     "",
			expectedLink:  "",
			expectChanged: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newLink, changed := r.relativize(tc.inputLink)
			assert.Equal(t, tc.expectedLink, newLink, "The rewritten link should match the expected value")
			assert.Equal(t, tc.expectChanged, changed, "The 'changed' flag should be as expected")
		})
	}
}
