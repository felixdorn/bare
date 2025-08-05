package url

import (
	"testing"
)

func TestPath_Matches(t *testing.T) {
	testCases := []struct {
		name    string
		pattern Path
		path    string
		want    bool
	}{
		// --- Exact Matches & Basic Wildcard ---
		{
			name:    "exact match",
			pattern: "/posts/a-post",
			path:    "/posts/a-post",
			want:    true,
		},
		{
			name:    "exact match with different slashes",
			pattern: "posts/a-post/",
			path:    "/posts/a-post",
			want:    true,
		},
		{
			name:    "no match",
			pattern: "/articles/*",
			path:    "/posts/my-post",
			want:    false,
		},

		// --- Single Star (*) Wildcard ---
		{
			name:    "simple wildcard match",
			pattern: "/posts/*",
			path:    "/posts/my-first-post",
			want:    true,
		},
		{
			name:    "wildcard match with numeric segment",
			pattern: "/posts/*",
			path:    "/posts/123",
			want:    true,
		},
		{
			name:    "wildcard mismatch with too many path segments",
			pattern: "/posts/*",
			path:    "/posts/a/b",
			want:    false,
		},
		{
			name:    "wildcard mismatch with too few path segments",
			pattern: "/posts/*",
			path:    "/posts/",
			want:    false,
		},
		{
			name:    "wildcard at start",
			pattern: "*/posts",
			path:    "anything/posts",
			want:    true,
		},
		{
			name:    "wildcard in middle",
			pattern: "/posts/*/comments",
			path:    "/posts/123/comments",
			want:    true,
		},
		{
			name:    "mismatch with wildcard in middle",
			pattern: "/posts/*/comments",
			path:    "/posts/123/author",
			want:    false,
		},

		// --- Double Star (**) Wildcard ---
		{
			name:    "double star at end matches one level",
			pattern: "/internal/**",
			path:    "/internal/page",
			want:    true,
		},
		{
			name:    "double star at end matches multiple levels",
			pattern: "/internal/**",
			path:    "/internal/sub/page",
			want:    true,
		},
		{
			name:    "double star at end matches zero levels (trailing slash)",
			pattern: "/internal/**",
			path:    "/internal/",
			want:    true,
		},
		{
			name:    "double star at end matches zero levels (no trailing slash)",
			pattern: "/internal/**",
			path:    "/internal",
			want:    true,
		},
		{
			name:    "double star does not match partial segment",
			pattern: "/internal/**",
			path:    "/internal-affairs/page",
			want:    false,
		},
		{
			name:    "double star at start",
			pattern: "**/secret.html",
			path:    "/api/v1/secret.html",
			want:    true,
		},
		{
			name:    "double star at start matches root",
			pattern: "**/secret.html",
			path:    "/secret.html",
			want:    true,
		},
		{
			name:    "double star in the middle",
			pattern: "/api/**/data",
			path:    "/api/v1/users/data",
			want:    true,
		},
		{
			name:    "double star in the middle matches zero segments",
			pattern: "/api/**/data",
			path:    "/api/data",
			want:    true,
		},
		{
			name:    "double star mismatch",
			pattern: "/api/**/data",
			path:    "/api/v1/users/metadata",
			want:    false,
		},
		{
			name:    "double star matches everything",
			pattern: "**",
			path:    "/any/thing/at/all",
			want:    true,
		},
		{
			name:    "double star alone matches root",
			pattern: "**",
			path:    "/",
			want:    true,
		},
		{
			name:    "complex pattern with multiple double stars",
			pattern: "/a/**/b/**/c",
			path:    "/a/x/y/b/z/c",
			want:    true,
		},
		{
			name:    "complex pattern with adjacent double stars",
			pattern: "/a/**/**/c",
			path:    "/a/b/c",
			want:    true,
		},
		{
			name:    "pattern ends with double star but path is shorter",
			pattern: "/a/b/**",
			path:    "/a",
			want:    false,
		},

		// --- Edge Cases ---
		{
			name:    "root path match",
			pattern: "/",
			path:    "/",
			want:    true,
		},
		{
			name:    "root pattern vs empty path",
			pattern: "/",
			path:    "",
			want:    true,
		},
		{
			name:    "root path mismatch",
			pattern: "/",
			path:    "/posts",
			want:    false,
		},
		{
			name:    "empty pattern vs root path",
			pattern: "",
			path:    "/",
			want:    true,
		},
		{
			name:    "both empty",
			pattern: "",
			path:    "",
			want:    true,
		},
		{
			name:    "only double star matches empty",
			pattern: "**",
			path:    "",
			want:    true,
		},
		{
			name:    "only single star matches empty",
			pattern: "*",
			path:    "",
			want:    true,
		},
		{
			name:    "pattern with content does not match empty",
			pattern: "a/b",
			path:    "",
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pattern.Matches(tc.path)
			if got != tc.want {
				t.Errorf("Path(%q).Matches(%q) = %v; want %v", tc.pattern, tc.path, got, tc.want)
			}
		})
	}
}
