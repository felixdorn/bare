package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/felixdorn/bare/core/domain/url"
	"golang.org/x/net/html"
)

type Page struct {
	Code  int
	URL   *url.URL
	Links []*url.URL
	Body  io.Reader
}

// Client is a wrapper around http.Client that provides page-fetching capabilities.
type Client struct {
	httpClient *http.Client
}

// New creates a new Client with a default timeout.
// It allows providing a custom *http.Client, which is useful for testing.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return &Client{
		httpClient: httpClient,
	}
}

// GetPage fetches a URL and parses it into a Page struct.
// It extracts all internal links from the page.
func (c *Client) GetPage(ctx context.Context, pageURL *url.URL) (*Page, error) {
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
		Code:  resp.StatusCode,
		URL:   pageURL,
		Body:  bytes.NewReader(bodyBytes),
		Links: []*url.URL{},
	}

	// Use a separate reader for parsing so we don't consume the one in the Page struct.
	z := html.NewTokenizer(bytes.NewReader(bodyBytes))

	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			// End of document
			break
		}

		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		t := z.Token()

		for _, attr := range t.Attr {
			// We're interested in attributes that can contain URLs.
			if attr.Key == "href" || attr.Key == "src" {
				linkURL, err := url.Parse(attr.Val)
				if err != nil {
					// Malformed URL, just ignore it and break to the next token
					break
				}

				// Only consider http and https links, or relative links.
				if linkURL.Scheme != "" && linkURL.Scheme != "http" && linkURL.Scheme != "https" {
					break
				}

				resolvedURL := pageURL.ResolveReference(linkURL)

				// Only follow links that are internal to the site.
				if resolvedURL.IsInternal(pageURL) {
					page.Links = append(page.Links, resolvedURL)
				}
			}
		}
	}

	return page, nil
}
