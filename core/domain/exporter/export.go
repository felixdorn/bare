package exporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/domain/crawler"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
)

// Common errors for link filtering
var (
	ErrExternal = errors.New("external link")
	ErrExcluded = errors.New("excluded by config")
)

// Export manages the exporting process using the crawler.
type Export struct {
	Conf       *config.Config
	log        zerolog.Logger
	httpClient *http.Client
}

// NewExport creates a new Export instance.
// The httpClient parameter is optional; if nil, a default client will be used.
func NewExport(conf *config.Config, log zerolog.Logger, httpClient *http.Client) *Export {
	return &Export{
		Conf:       conf,
		log:        log,
		httpClient: httpClient,
	}
}

// Run executes the export process using the crawler.
func (e *Export) Run(ctx context.Context) error {
	// Convert entrypoints to strings
	entrypoints := make([]string, len(e.Conf.Pages.Entrypoints))
	for i, ep := range e.Conf.Pages.Entrypoints {
		entrypoints[i] = string(ep)
	}

	c := crawler.New(crawler.Config{
		BaseURL:     e.Conf.URL,
		WorkerCount: e.Conf.WorkersCount,
		Entrypoints: entrypoints,
		Logger:      e.log,
		HTTPClient:  e.httpClient,

		OnNewLink: func(page *crawler.Page, link crawler.Link) error {
			// Only extract links from crawlable pages (HTML)
			if !isCrawlable(page.URL) {
				return errors.New("source page is not crawlable")
			}

			// Only follow internal links
			if !link.URL.IsInternal(page.URL) {
				return ErrExternal
			}

			// Check if URL is allowed
			if !e.Conf.IsURLAllowed(link.URL) {
				return ErrExcluded
			}

			return nil
		},

		OnPage: func(page *crawler.Page) {
			// If the page is marked as extract-only, don't save it
			if e.Conf.IsExtractOnly(page.URL) {
				e.log.Info().Str("url", page.URL.String()).Msg("Extracting links only, skipping save")
				return
			}

			// Save the page to disk
			if err := e.savePage(page); err != nil {
				e.log.Error().Err(err).Str("url", page.URL.String()).Msg("Failed to save page")
			}
		},
	})

	err := c.Run(ctx)

	if ctx.Err() != nil {
		fmt.Println("\nExport cancelled.")
		return nil
	}

	if err != nil {
		return err
	}

	fmt.Println("Export finished.")
	return nil
}

// savePage writes a page's content to disk.
func (e *Export) savePage(page *crawler.Page) error {
	path := page.URL.ToPath(e.Conf.Output)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(page.Body))
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	e.log.Info().Str("url", page.URL.String()).Str("path", path).Msg("Exported page")
	return nil
}

// isCrawlable checks if a URL should be crawled for more links.
func isCrawlable(u *url.URL) bool {
	ext := filepath.Ext(u.Path)
	return ext == "" || ext == ".html"
}
