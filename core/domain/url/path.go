package url

import (
	"strings"
)

// Path represents a URL path pattern for matching against other paths.
type Path string

// Match checks if the given path string `s` matches the pattern `p`.
// The matching is path-segment based and supports wildcards.
// It ignores leading and trailing slashes for both the pattern and the path.
//
// Wildcards:
//   - `*`: matches any single path segment.
//   - `**`: matches zero or more path segments.
//
// Examples:
//   - Pattern "/posts/*" matches "/posts/my-first-post" but not "/posts/a/b" or "/posts/".
//   - Pattern "/posts/**" matches "/posts/my-first-post", "/posts/a/b", and "/posts/".
//   - Pattern "/**/secret" matches "/top/secret" and "/secret".
func (p Path) Matches(s string) bool {
	pattern := strings.Trim(string(p), "/")
	path := strings.Trim(s, "/")

	pSegs := strings.Split(pattern, "/")
	sSegs := strings.Split(path, "/")

	pIdx, sIdx := 0, 0
	// starIdx stores the position of the last `**` in pSegs
	// sTmpIdx stores the position in sSegs that we're trying to match from
	starIdx, sTmpIdx := -1, -1

	for sIdx < len(sSegs) {
		// Case 1: Segments match (or pattern has '*')
		if pIdx < len(pSegs) && (pSegs[pIdx] == "*" || pSegs[pIdx] == sSegs[sIdx]) {
			pIdx++
			sIdx++
			continue
		}

		// Case 2: Pattern has '**'. It can match the current segment.
		// We save its position and the current path position, then advance the pattern pointer.
		if pIdx < len(pSegs) && pSegs[pIdx] == "**" {
			starIdx = pIdx
			sTmpIdx = sIdx
			pIdx++
			continue
		}

		// Case 3: No match, but we have a previous '**' to backtrack to.
		// We reset the pattern pointer to after the '**' and advance the saved path pointer.
		// This makes the `**` consume one more segment from the path.
		if starIdx != -1 {
			pIdx = starIdx + 1
			sTmpIdx++
			sIdx = sTmpIdx
			continue
		}

		// Case 4: No match and no `**` to backtrack to.
		return false
	}

	// The path is exhausted. The rest of the pattern must be only `**`s,
	// which can match an empty sequence.
	for pIdx < len(pSegs) && pSegs[pIdx] == "**" {
		pIdx++
	}

	return pIdx == len(pSegs)
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
