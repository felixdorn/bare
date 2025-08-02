package url

import (
	"strings"
)

// Path represents a URL path pattern for matching against other paths.
type Path string

// Match checks if the given path string `s` matches the pattern `p`.
// The matching is path-segment based. A `*` in the pattern matches any single segment.
// It ignores leading and trailing slashes for both the pattern and the path.
//
// Examples:
//   - Pattern "/posts/*" will match "/posts/my-first-post" and "posts/123".
//   - Pattern "/posts/*" will NOT match "/posts", "/posts/a/b", or "/articles/my-first-post".
//   - Pattern "/" will match "/".
func (p Path) Matches(s string) bool {
	pattern := strings.Trim(string(p), "/")
	path := strings.Trim(s, "/")

	// strings.Split on an empty string (like the root path "/" after trimming)
	// produces a slice with a single empty string: `[""]`.
	// This correctly handles matching "/" with "/".
	patternSegments := strings.Split(pattern, "/")
	pathSegments := strings.Split(path, "/")

	if len(patternSegments) != len(pathSegments) {
		return false
	}

	for i, segment := range patternSegments {
		if segment == "*" {
			continue
		}
		if segment != pathSegments[i] {
			return false
		}
	}

	return true
}

// Paths is a collection of Path patterns.
type Paths []Path

// MatchAny checks if the given path string `s` matches any of the patterns in the collection.
func (ps Paths) MatchAny(s string) bool {
	for _, p := range ps {
		if p.Matches(s) {
			return true
		}
	}
	return false
}

// String provides a string representation of Paths for logging.
func (ps Paths) String() string {
	if ps == nil {
		return ""
	}
	paths := make([]string, len(ps))
	for i, p := range ps {
		paths[i] = string(p)
	}
	return strings.Join(paths, ", ")
}
