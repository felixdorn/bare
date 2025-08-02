package url

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// URL is a wrapper around the standard `url.URL` type.
type URL struct {
	*url.URL
}

// Parse parses a raw url string into a URL structure.
func Parse(rawURL string) (*URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse url %s: %w", rawURL, err)
	}

	return &URL{u}, nil
}

// IsInternal checks if the URL is internal to a given host.
// It considers a URL internal if it has the same host and does not handle subdomains.
func (u *URL) IsInternal(baseURL *URL) bool {
	if u.Hostname() == "" {
		return true
	}
	return u.Hostname() == baseURL.Hostname()
}

// ToPath converts the URL's path to a file system path.
// It appends 'index.html' to paths that end with a '/' or have no extension.
func (u *URL) ToPath(root string) string {
	path := filepath.Join(root, u.Path)

	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	} else if filepath.Ext(path) == "" {
		// This is a rough check for extension-less URLs that should be treated as directories.
		path = filepath.Join(path, "index.html")
	}
	return path
}

// ResolveReference resolves a reference URL against the base URL.
func (u *URL) ResolveReference(ref *URL) *URL {
	return &URL{u.URL.ResolveReference(ref.URL)}
}

// String returns the string representation of the URL.
func (u *URL) String() string {
	if u == nil || u.URL == nil {
		return ""
	}
	return u.URL.String()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (u *URL) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))
	if err != nil {
		return err
	}
	u.URL = parsed
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (u *URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}
