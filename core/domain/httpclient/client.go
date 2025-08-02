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

// httpClient is the package-level HTTP client with a timeout.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func GetPage(ctx context.Context, pageURL *url.URL) (*Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	page := &Page{
		Code:  resp.StatusCode,
		URL:   pageURL,
		Body:  bytes.NewReader(byt),
		Links: []*url.URL{},
	}

	// Use a separate reader for parsing so we don't consume the one in the Page struct.
	z := html.NewTokenizer(bytes.NewReader(byt))

	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			break
		}

		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		t := z.Token()

		for _, attr := range t.Attr {
			if attr.Key == "href" || attr.Key == "src" {
				linkURL, err := url.Parse(attr.Val)
				if err != nil {
					// Malformed URL, just ignore it and break to the next token
					break
				}

				// Only process http and https links.
				if linkURL.Scheme != "" && linkURL.Scheme != "http" && linkURL.Scheme != "https" {
					break
				}

				resolvedURL := pageURL.ResolveReference(linkURL)
				if resolvedURL.IsInternal(pageURL) {
					page.Links = append(page.Links, resolvedURL)
				}
				break
			}
		}
	}

	return page, nil
}
