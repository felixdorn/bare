package config

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_IsURLAllowed(t *testing.T) {
	baseURL, _ := url.Parse("http://example.com")

	testCases := []struct {
		name     string
		config   *Config
		urlPath  string
		expected bool
	}{
		{
			name: "no rules, should allow",
			config: &Config{
				URL: baseURL,
				Pages: Pages{
					Exclude: url.Paths{},
				},
			},
			urlPath:  "/any/path",
			expected: true,
		},
		{
			name: "exclude rule matches, should deny",
			config: &Config{
				URL: baseURL,
				Pages: Pages{
					Exclude: url.Paths{"/private/*"},
				},
			},
			urlPath:  "/private/page",
			expected: false,
		},
		{
			name: "exclude rule does not match, should allow",
			config: &Config{
				URL: baseURL,
				Pages: Pages{
					Exclude: url.Paths{"/private/*"},
				},
			},
			urlPath:  "/public/page",
			expected: true,
		},
		{
			name: "root path excluded",
			config: &Config{
				URL: baseURL,
				Pages: Pages{
					Exclude: url.Paths{"/"},
				},
			},
			urlPath:  "/",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(baseURL.String() + tc.urlPath)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, tc.config.IsURLAllowed(u))
		})
	}
}
