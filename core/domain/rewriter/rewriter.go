package rewriter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/felixdorn/bare/core/domain/url"

	"golang.org/x/net/html"
)

// Rewriter walks a directory of static files and rewrites absolute URLs
// to be root-relative, making the entire site self-contained.
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

// Run executes the rewriting process. It walks the output directory,
// finds HTML files, and rewrites their internal links to be root-relative.
func (r *Rewriter) Run() error {
	absOutputDir, err := filepath.Abs(r.OutputDir)
	if err != nil {
		return err
	}

	return filepath.Walk(absOutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".html") {
			return r.rewriteFile(path)
		}

		return nil
	})
}

// rewriteFile reads an HTML file, rewrites its links, and saves it back to disk if changes were made.
func (r *Rewriter) rewriteFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return err
	}

	changed := false
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				// We are interested in attributes that typically contain URLs.
				if attr.Key == "href" || attr.Key == "src" {
					newVal, wasChanged := r.relativize(attr.Val)
					if wasChanged {
						n.Attr[i].Val = newVal
						changed = true
					}
				}
			}
		}

		// Recursively visit all child nodes.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}

	visit(doc)

	// Only write the file back if it has been modified.
	if changed {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Render the modified HTML tree back to the file.
		return html.Render(file, doc)
	}

	return nil
}

// relativize converts an absolute URL into a root-relative path.
// It returns the potentially modified URL and a boolean indicating if a change was made.
func (r *Rewriter) relativize(link string) (string, bool) {
	linkURL, err := url.Parse(link)
	if err != nil {
		// Ignore malformed URLs.
		return link, false
	}

	// We only want to process absolute URLs that point to our own site.
	// An absolute URL must have a scheme and a host.
	// - Must be http or https
	// - Hostname must not be empty and must match our base URL's hostname
	if (linkURL.Scheme != "http" && linkURL.Scheme != "https") ||
		linkURL.Hostname() == "" ||
		linkURL.Hostname() != r.BaseURL.Hostname() {
		return link, false
	}

	// Construct the new root-relative path.
	newPath := linkURL.Path
	if newPath == "" {
		newPath = "/"
	}

	// Preserve query parameters and fragments.
	if linkURL.RawQuery != "" {
		newPath += "?" + linkURL.RawQuery
	}
	if linkURL.Fragment != "" {
		newPath += "#" + linkURL.Fragment
	}

	return newPath, true
}
