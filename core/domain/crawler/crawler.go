package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
	"golang.org/x/net/html"
)

// Link represents a link found on a page.
type Link struct {
	URL  *url.URL
	Text string // anchor text
	Rel  string // rel attribute (e.g., "nofollow", "noopener")
}

// Page represents a crawled page with its metadata and content.
type Page struct {
	URL         *url.URL
	StatusCode  int
	Body        []byte // raw HTML
	Links       []Link // ALL links (internal + external)
	Title       string
	Description string
	Canonical   string
}

// Config holds the crawler configuration.
type Config struct {
	BaseURL     *url.URL
	WorkerCount int
	Entrypoints []string
	Logger      zerolog.Logger
	HTTPClient  *http.Client

	// OnNewLink is called for every link discovered on a page.
	// Return nil to follow the link, or an error to skip it.
	// The error is used for logging/debugging purposes.
	OnNewLink func(page *Page, link Link) error

	// OnPage is called when a page has been fully crawled.
	OnPage func(page *Page)
}

// Crawler manages the crawling process.
type Crawler struct {
	cfg        Config
	httpClient *http.Client
	log        zerolog.Logger
}

// workerResult holds the outcome of a worker's task.
type workerResult struct {
	pageURL *url.URL
	page    *Page
	toQueue []*url.URL // URLs that passed OnNewLink filter
	err     error
}

// normalizeURL returns a canonical string representation of a URL for deduplication.
// It strips fragments and normalizes the scheme to match the base URL.
func normalizeURL(u *url.URL, baseURL *url.URL) string {
	// Create a copy to avoid modifying the original
	normalized := *u.URL
	// Strip fragment
	normalized.Fragment = ""
	// Normalize scheme to match base URL
	normalized.Scheme = baseURL.Scheme
	return normalized.String()
}

// New creates a new Crawler instance.
func New(cfg Config) *Crawler {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	workerCount := cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = 10
	}
	cfg.WorkerCount = workerCount

	return &Crawler{
		cfg:        cfg,
		httpClient: httpClient,
		log:        cfg.Logger,
	}
}

// Run executes the crawling process.
func (c *Crawler) Run(ctx context.Context) error {
	numWorkers := c.cfg.WorkerCount
	c.log.Debug().Msg("Starting crawler with centralized controller")

	tasksChan := make(chan *url.URL, numWorkers)
	resultsChan := make(chan workerResult, numWorkers)
	var workersWg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go c.worker(ctx, i+1, &workersWg, tasksChan, resultsChan)
	}
	c.log.Debug().Int("workers", numWorkers).Msg("Started workers")

	// Controller state
	queue := make([]*url.URL, 0)
	visited := make(map[string]bool)
	activeWorkers := 0

	// Seed the queue with initial entrypoints
	for _, p := range c.cfg.Entrypoints {
		u, err := url.Parse(p)
		if err != nil {
			c.log.Error().Err(err).Str("path", p).Msg("Invalid entrypoint path")
			continue
		}
		resolvedURL := c.cfg.BaseURL.ResolveReference(u)
		normalizedKey := normalizeURL(resolvedURL, c.cfg.BaseURL)

		if !visited[normalizedKey] {
			visited[normalizedKey] = true
			queue = append(queue, resolvedURL)
		}
	}
	c.log.Debug().Int("queue_size", len(queue)).Msg("Initial queue populated")

	// Controller loop
controllerLoop:
	for len(queue) > 0 || activeWorkers > 0 {
		var task *url.URL
		var sendChan chan<- *url.URL

		if len(queue) > 0 {
			task = queue[0]
			sendChan = tasksChan
		}

		select {
		case <-ctx.Done():
			c.log.Info().Msg("Context cancelled, shutting down...")
			break controllerLoop

		case sendChan <- task:
			c.log.Debug().Str("url", task.String()).Msg("Sent task")
			queue = queue[1:]
			activeWorkers++

		case result := <-resultsChan:
			c.log.Debug().Str("url", result.pageURL.String()).Msg("Received result")
			activeWorkers--

			if result.err != nil {
				if ctx.Err() == nil {
					c.log.Error().Err(result.err).Str("url", result.pageURL.String()).Msg("Failed to process URL")
				}
				continue
			}

			// Call OnPage callback
			if c.cfg.OnPage != nil && result.page != nil {
				c.cfg.OnPage(result.page)
			}

			// Add URLs that passed the OnNewLink filter to the queue
			for _, link := range result.toQueue {
				normalizedKey := normalizeURL(link, c.cfg.BaseURL)
				if !visited[normalizedKey] {
					visited[normalizedKey] = true
					queue = append(queue, link)
					c.log.Debug().Str("url", link.String()).Msg("Queued new link")
				}
			}
		}
	}

	// Graceful shutdown
	c.log.Debug().Msg("Work finished or cancelled. Starting cleanup.")

	for activeWorkers > 0 {
		<-resultsChan
		activeWorkers--
		c.log.Debug().Int("remaining_workers", activeWorkers).Msg("Draining worker result.")
	}

	c.log.Debug().Msg("Closing tasks channel.")
	close(tasksChan)

	c.log.Debug().Msg("Waiting for all workers to terminate.")
	workersWg.Wait()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

// worker fetches URLs and processes them.
func (c *Crawler) worker(ctx context.Context, id int, wg *sync.WaitGroup, tasks <-chan *url.URL, results chan<- workerResult) {
	defer wg.Done()
	log := c.log.With().Int("worker_id", id).Logger()
	log.Debug().Msg("Worker started")

	for pageURL := range tasks {
		log.Debug().Str("url", pageURL.String()).Msg("Received task")

		page, err := c.fetchPage(ctx, pageURL)
		if err != nil {
			results <- workerResult{pageURL: pageURL, err: fmt.Errorf("failed to get page: %w", err)}
			continue
		}
		log.Debug().Str("url", pageURL.String()).Msg("Successfully fetched page")

		// Filter links through OnNewLink callback
		var toQueue []*url.URL
		if c.cfg.OnNewLink != nil {
			for _, link := range page.Links {
				if err := c.cfg.OnNewLink(page, link); err == nil {
					toQueue = append(toQueue, link.URL)
				} else {
					log.Debug().Str("url", link.URL.String()).Err(err).Msg("Link filtered out")
				}
			}
		}

		results <- workerResult{pageURL: pageURL, page: page, toQueue: toQueue}
	}

	log.Debug().Msg("Worker shutting down")
}

// fetchPage fetches a URL and parses it into a Page struct.
func (c *Crawler) fetchPage(ctx context.Context, pageURL *url.URL) (*Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	page := &Page{
		URL:        pageURL,
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
		Links:      []Link{},
	}

	// Parse HTML for metadata and links
	c.parseHTML(page, bodyBytes)

	return page, nil
}

// parseHTML extracts metadata and links from HTML content.
func (c *Crawler) parseHTML(page *Page, body []byte) {
	z := html.NewTokenizer(bytes.NewReader(body))

	var inTitle bool
	var currentLinkText strings.Builder

	// Track the current <a> tag we're processing
	var currentAnchor *Link

	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			break
		}

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			tagName := t.Data

			switch tagName {
			case "title":
				inTitle = true

			case "meta":
				name, content := "", ""
				for _, attr := range t.Attr {
					switch attr.Key {
					case "name":
						name = strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					}
				}
				if name == "description" {
					page.Description = content
				}

			case "link":
				rel, href := "", ""
				for _, attr := range t.Attr {
					switch attr.Key {
					case "rel":
						rel = attr.Val
					case "href":
						href = attr.Val
					}
				}
				if rel == "canonical" && href != "" {
					page.Canonical = href
				}
				// Also extract stylesheet and other link types as links
				if href != "" && rel != "canonical" {
					linkURL, err := url.Parse(href)
					if err == nil && (linkURL.Scheme == "" || linkURL.Scheme == "http" || linkURL.Scheme == "https") {
						resolvedURL := page.URL.ResolveReference(linkURL)
						page.Links = append(page.Links, Link{
							URL:  resolvedURL,
							Text: "",
							Rel:  rel,
						})
					}
				}

			case "a":
				href, rel := "", ""
				for _, attr := range t.Attr {
					switch attr.Key {
					case "href":
						href = attr.Val
					case "rel":
						rel = attr.Val
					}
				}
				if href != "" {
					linkURL, err := url.Parse(href)
					if err == nil && (linkURL.Scheme == "" || linkURL.Scheme == "http" || linkURL.Scheme == "https") {
						resolvedURL := page.URL.ResolveReference(linkURL)
						currentAnchor = &Link{
							URL: resolvedURL,
							Rel: rel,
						}
						currentLinkText.Reset()
					}
				}

			default:
				// Handle other elements with src attribute (images, scripts, etc.)
				for _, attr := range t.Attr {
					if attr.Key == "src" {
						linkURL, err := url.Parse(attr.Val)
						if err == nil && (linkURL.Scheme == "" || linkURL.Scheme == "http" || linkURL.Scheme == "https") {
							resolvedURL := page.URL.ResolveReference(linkURL)
							page.Links = append(page.Links, Link{
								URL:  resolvedURL,
								Text: "", // src elements don't have anchor text
								Rel:  "",
							})
						}
						break
					}
				}
			}

		case html.EndTagToken:
			t := z.Token()
			switch t.Data {
			case "title":
				inTitle = false
			case "a":
				if currentAnchor != nil {
					currentAnchor.Text = strings.TrimSpace(currentLinkText.String())
					page.Links = append(page.Links, *currentAnchor)
					currentAnchor = nil
				}
			}

		case html.TextToken:
			text := string(z.Text())
			if inTitle {
				page.Title = strings.TrimSpace(text)
			}
			if currentAnchor != nil {
				currentLinkText.WriteString(text)
			}
		}
	}
}
