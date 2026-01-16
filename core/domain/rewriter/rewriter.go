package rewriter

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/felixdorn/bare/core/domain/url"
)

// Rewriter walks a directory of exported files and rewrites absolute URLs
// to be root-relative when the target file exists in the export.
type Rewriter struct {
	OutputDir string
	BaseURL   *url.URL
	urlRegex  *regexp.Regexp
}

// New creates a new Rewriter instance.
func New(outputDir string, baseURL *url.URL) *Rewriter {
	// Build regex to match URLs with this base
	// Matches: http(s)://hostname/path with optional query and fragment
	escaped := regexp.QuoteMeta(baseURL.Scheme + "://" + baseURL.Host)
	pattern := escaped + `(/[^\s"'<>)*]*)?`

	return &Rewriter{
		OutputDir: outputDir,
		BaseURL:   baseURL,
		urlRegex:  regexp.MustCompile(pattern),
	}
}

// Run executes the rewriting process on all files in the output directory.
func (r *Rewriter) Run() error {
	absOutputDir, err := filepath.Abs(r.OutputDir)
	if err != nil {
		return err
	}

	return filepath.Walk(absOutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return r.rewriteFile(path, absOutputDir)
	})
}

// rewriteFile processes a single file, rewriting URLs where the target exists.
func (r *Rewriter) rewriteFile(filePath, outputDir string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	original := string(content)
	result := r.urlRegex.ReplaceAllStringFunc(original, func(match string) string {
		return r.processURL(match, outputDir)
	})

	// Only write if changed
	if result != original {
		return os.WriteFile(filePath, []byte(result), 0644)
	}

	return nil
}

// processURL decides whether to rewrite a URL and how.
func (r *Rewriter) processURL(rawURL, outputDir string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Check for norewrite param - leave URL as-is
	if parsed.Query().Has("norewrite") {
		return rawURL
	}

	// Get the path portion
	path := parsed.Path
	if path == "" {
		path = "/"
	}

	// Check if target exists in the export
	if !r.targetExists(path, outputDir) {
		return rawURL
	}

	// Rewrite to root-relative
	result := path
	if parsed.RawQuery != "" {
		result += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		result += "#" + parsed.Fragment
	}

	return result
}

// targetExists checks if a URL path corresponds to an existing file in the export.
func (r *Rewriter) targetExists(urlPath, outputDir string) bool {
	// Convert URL path to filesystem path
	fsPath := filepath.Join(outputDir, filepath.FromSlash(urlPath))

	// Check if it exists directly
	if info, err := os.Stat(fsPath); err == nil {
		if !info.IsDir() {
			return true
		}
		// If it's a directory, check for index.html
		indexPath := filepath.Join(fsPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return true
		}
	}

	// For paths without extension, try adding .html
	if filepath.Ext(fsPath) == "" && !strings.HasSuffix(fsPath, "/") {
		if _, err := os.Stat(fsPath + ".html"); err == nil {
			return true
		}
	}

	return false
}
