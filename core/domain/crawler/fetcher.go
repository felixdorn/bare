package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/felixdorn/bare/core/domain/url"
)

// FetchResult contains the raw response from fetching a URL.
type FetchResult struct {
	StatusCode int
	Body       []byte
}

// Fetcher abstracts how pages are fetched.
// Implementations can use HTTP, headless Chrome, or other methods.
type Fetcher interface {
	// Fetch retrieves the content at the given URL.
	Fetch(ctx context.Context, u *url.URL) (*FetchResult, error)

	// Close releases any resources held by the fetcher.
	// For HTTPFetcher this is a no-op, for JSFetcher it kills the Chrome process.
	Close() error
}

// HTTPFetcher fetches pages using a standard HTTP client.
type HTTPFetcher struct {
	client *http.Client
}

// NewHTTPFetcher creates a new HTTPFetcher.
// If client is nil, a default client with 10 second timeout is used.
func NewHTTPFetcher(client *http.Client) *HTTPFetcher {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return &HTTPFetcher{client: client}
}

// Fetch retrieves the content at the given URL using HTTP GET.
func (f *HTTPFetcher) Fetch(ctx context.Context, u *url.URL) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %w", u, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	return &FetchResult{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

// Close is a no-op for HTTPFetcher.
func (f *HTTPFetcher) Close() error {
	return nil
}
