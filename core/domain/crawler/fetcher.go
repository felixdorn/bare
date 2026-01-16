package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/felixdorn/bare/core/domain/url"
)

// Redirect represents a single redirect in a chain.
type Redirect struct {
	URL        string
	StatusCode int
}

// FetchResult contains the raw response from fetching a URL.
type FetchResult struct {
	StatusCode    int
	Body          []byte
	RedirectChain []Redirect // Ordered list of redirects (empty if no redirects)
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
	timeout time.Duration
}

// NewHTTPFetcher creates a new HTTPFetcher.
// If client is nil, a default client with 10 second timeout is used.
func NewHTTPFetcher(client *http.Client) *HTTPFetcher {
	timeout := 10 * time.Second
	if client != nil && client.Timeout > 0 {
		timeout = client.Timeout
	}
	return &HTTPFetcher{timeout: timeout}
}

// Fetch retrieves the content at the given URL using HTTP GET.
func (f *HTTPFetcher) Fetch(ctx context.Context, u *url.URL) (*FetchResult, error) {
	var chain []Redirect

	client := &http.Client{
		Timeout: f.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				prev := via[len(via)-1]
				chain = append(chain, Redirect{
					URL:        prev.URL.String(),
					StatusCode: prev.Response.StatusCode,
				})
			}
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %w", u, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	return &FetchResult{
		StatusCode:    resp.StatusCode,
		Body:          body,
		RedirectChain: chain,
	}, nil
}

// Close is a no-op for HTTPFetcher.
func (f *HTTPFetcher) Close() error {
	return nil
}
