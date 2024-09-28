package httpclient

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
)

type Page struct {
	Code  int
	URL   *url.URL
	Links []string
	Body  io.Reader
}

func GetPage(URL string) (*Page, error) {
	parsedURL, err := url.ParseRequestURI(URL)
	if err != nil {
		return nil, fmt.Errorf("could not parse Value: %w", err)
	}

	resp, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %w", URL, err)
	}

	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	resp.Body.Close()

	page := &Page{
		Code: resp.StatusCode,
		URL:  parsedURL,
		Body: bytes.NewReader(byt),
	}

	z := html.NewTokenizer(page.Body)

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
				page.Links = append(page.Links, attr.Val)
				break
			}
		}
	}

	return page, nil
}
