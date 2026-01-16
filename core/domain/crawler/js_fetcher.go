package crawler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
)

// JSFetcherOptions configures the JSFetcher.
type JSFetcherOptions struct {
	Wait           int      // milliseconds to wait for JS execution
	MaxTabs        int      // max parallel Chrome tabs (default 1 = sequential)
	ExecutablePath string   // path to Chrome/Chromium executable
	Flags          []string // additional Chrome flags
	Logger         zerolog.Logger
}

// JSFetcher fetches pages using headless Chrome, allowing JavaScript to execute.
type JSFetcher struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	opts     JSFetcherOptions
	sem      chan struct{} // semaphore for tab limiting
}

// NewJSFetcher creates a new JSFetcher that uses headless Chrome.
func NewJSFetcher(opts JSFetcherOptions) (*JSFetcher, error) {
	if opts.MaxTabs <= 0 {
		opts.MaxTabs = 1
	}
	if opts.Wait <= 0 {
		opts.Wait = 2000
	}

	allocatorOptions := chromedp.DefaultExecAllocatorOptions[:]

	if opts.ExecutablePath != "" {
		allocatorOptions = append(allocatorOptions, chromedp.ExecPath(opts.ExecutablePath))
	}

	for _, flag := range opts.Flags {
		flag = strings.TrimPrefix(flag, "--")
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) == 1 {
			allocatorOptions = append(allocatorOptions, chromedp.Flag(parts[0], true))
		} else {
			allocatorOptions = append(allocatorOptions, chromedp.Flag(parts[0], parts[1]))
		}
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)

	opts.Logger.Info().Msg("Starting headless Chrome for JS fetching")

	return &JSFetcher{
		allocCtx: allocCtx,
		cancel:   cancel,
		opts:     opts,
		sem:      make(chan struct{}, opts.MaxTabs),
	}, nil
}

// Fetch navigates to the URL, waits for JS to execute, and returns the rendered HTML.
func (f *JSFetcher) Fetch(ctx context.Context, u *url.URL) (*FetchResult, error) {
	// Acquire semaphore slot
	select {
	case f.sem <- struct{}{}:
		defer func() { <-f.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Create a new browser context (tab) for this fetch
	var browserOpts []chromedp.ContextOption
	if f.opts.Logger.GetLevel() == zerolog.DebugLevel {
		browserOpts = append(browserOpts, chromedp.WithLogf(f.opts.Logger.Printf))
	}

	taskCtx, cancel := chromedp.NewContext(f.allocCtx, browserOpts...)
	defer cancel()

	// Track the response status code
	var statusCode int
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		if resp, ok := ev.(*network.EventResponseReceived); ok {
			// Only track the main document response
			if resp.Type == network.ResourceTypeDocument {
				statusCode = int(resp.Response.Status)
			}
		}
	})

	var html string
	err := chromedp.Run(taskCtx,
		network.Enable(),
		chromedp.Navigate(u.String()),
		chromedp.Sleep(time.Duration(f.opts.Wait)*time.Millisecond),
		chromedp.Evaluate(`document.documentElement.outerHTML`, &html),
	)
	if err != nil {
		return nil, fmt.Errorf("chrome fetch failed for %s: %w", u, err)
	}

	// Default to 200 if we didn't capture a status
	if statusCode == 0 {
		statusCode = 200
	}

	return &FetchResult{
		StatusCode: statusCode,
		Body:       []byte(html),
	}, nil
}

// Close shuts down the Chrome process.
func (f *JSFetcher) Close() error {
	f.cancel()
	return nil
}
