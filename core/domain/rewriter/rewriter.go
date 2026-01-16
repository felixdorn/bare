package rewriter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/felixdorn/bare/core/domain/url"
)

// Rewriter walks a directory of exported files and rewrites absolute URLs
// to be root-relative, making the entire site self-contained.
//
// The rewriter only rewrites URLs that point to files that exist in the export.
// This means external URLs and URLs to missing files are left unchanged.
//
// To prevent a specific URL from being rewritten, add ?norewrite to it:
//
//	http://example.com/page?norewrite → http://example.com/page (kept absolute)
//	http://example.com/page           → /page (rewritten to relative)
//
// The ?norewrite parameter is removed from the final output.
type Rewriter struct {
	OutputDir string
	BaseURL   *url.URL
}

// New creates a new Rewriter instance.
func New(outputDir string, baseURL *url.URL) *Rewriter {
	return &Rewriter{
		OutputDir: outputDir,
		BaseURL:   baseURL,
	}
}

// Run executes the rewriting process. It builds an index of all exported files,
// then rewrites any absolute URLs that point to those files.
func (r *Rewriter) Run() error {
	absOutputDir, err := filepath.Abs(r.OutputDir)
	if err != nil {
		return err
	}

	// Build index of all exported paths
	exportedPaths := make(map[string]bool)
	err = filepath.Walk(absOutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(absOutputDir, path)
			// Store as URL path (forward slashes, leading /)
			urlPath := "/" + filepath.ToSlash(relPath)
			exportedPaths[urlPath] = true

			// Also index without index.html suffix for directory URLs
			if strings.HasSuffix(urlPath, "/index.html") {
				dirPath := strings.TrimSuffix(urlPath, "index.html")
				exportedPaths[dirPath] = true
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Build the base URL string to search for
	baseURLStr := r.BaseURL.Scheme + "://" + r.BaseURL.Host

	// Now rewrite all files
	return filepath.Walk(absOutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		return r.rewriteFile(path, baseURLStr, exportedPaths)
	})
}

// rewriteFile reads a file, replaces absolute URLs with relative ones, and saves if changed.
func (r *Rewriter) rewriteFile(filePath, baseURLStr string, exportedPaths map[string]bool) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	original := string(content)
	result := original

	// Find and replace all occurrences of the base URL
	// We look for the base URL followed by a path
	searchStart := 0
	for {
		idx := strings.Index(result[searchStart:], baseURLStr)
		if idx == -1 {
			break
		}
		idx += searchStart

		// Find the end of the URL (space, quote, >, or end of string)
		urlStart := idx
		urlEnd := idx + len(baseURLStr)

		// Extract the path portion
		for urlEnd < len(result) {
			ch := result[urlEnd]
			// Stop at URL terminators
			if ch == '"' || ch == '\'' || ch == '>' || ch == '<' || ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
				break
			}
			urlEnd++
		}

		fullURL := result[urlStart:urlEnd]
		pathPortion := fullURL[len(baseURLStr):]

		// Handle empty path as /
		if pathPortion == "" {
			pathPortion = "/"
		}

		// Split off query string and fragment for checking
		pathOnly := pathPortion
		queryAndFragment := ""
		if qIdx := strings.Index(pathOnly, "?"); qIdx != -1 {
			queryAndFragment = pathOnly[qIdx:]
			pathOnly = pathOnly[:qIdx]
		} else if fIdx := strings.Index(pathOnly, "#"); fIdx != -1 {
			queryAndFragment = pathOnly[fIdx:]
			pathOnly = pathOnly[:fIdx]
		}

		// Check for ?norewrite flag
		noRewrite := strings.Contains(queryAndFragment, "norewrite")
		if noRewrite {
			// Remove norewrite from query string
			queryAndFragment = strings.Replace(queryAndFragment, "?norewrite&", "?", 1)
			queryAndFragment = strings.Replace(queryAndFragment, "&norewrite", "", 1)
			queryAndFragment = strings.Replace(queryAndFragment, "?norewrite", "", 1)
		}

		// Check if this path exists in our export
		if exportedPaths[pathOnly] {
			var newURL string
			if noRewrite {
				// Keep absolute URL but remove norewrite param
				newURL = fullURL[:len(baseURLStr)] + pathOnly + queryAndFragment
			} else {
				// Replace with relative path
				newURL = pathOnly + queryAndFragment
			}
			result = result[:urlStart] + newURL + result[urlEnd:]
			searchStart = urlStart + len(newURL)
		} else {
			// Move past this URL
			searchStart = urlEnd
		}
	}

	// Only write if changed
	if result != original {
		return os.WriteFile(filePath, []byte(result), 0644)
	}

	return nil
}
