package js

import (
	"context"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
)

// Crawler is an interface for crawling a site with JavaScript.
type Crawler interface {
	Run(ctx context.Context) ([]*url.URL, error)
}

// JS is a crawler that uses chromedp to crawl a site.
type JS struct {
	config *config.Config
	log    zerolog.Logger

	urls     map[string]struct{}
	urlsLock sync.Mutex
}

// New creates a new JS crawler.
func New(config *config.Config, log zerolog.Logger) *JS {
	return &JS{
		config: config,
		log:    log,
		urls:   make(map[string]struct{}),
	}
}

// Run crawls the site and returns all found URLs.
func (j *JS) Run(ctx context.Context) ([]*url.URL, error) {
	j.log.Info().Msg("Starting JavaScript-based crawl to discover assets...")

	// create context
	allocatorOptions := chromedp.DefaultExecAllocatorOptions[:]
	if j.config.JS.ExecutablePath != "" {
		allocatorOptions = append(allocatorOptions, chromedp.ExecPath(j.config.JS.ExecutablePath))
	}

	for _, flag := range j.config.JS.Flags {
		flag = strings.TrimPrefix(flag, "--")
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) == 1 {
			allocatorOptions = append(allocatorOptions, chromedp.Flag(parts[0], true))
		} else {
			allocatorOptions = append(allocatorOptions, chromedp.Flag(parts[0], parts[1]))
		}
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)
	defer cancel()

	opts := []chromedp.ContextOption{}
	if j.log.GetLevel() == zerolog.DebugLevel {
		opts = append(opts, chromedp.WithLogf(j.log.Printf))
	}

	taskCtx, cancel := chromedp.NewContext(allocCtx, opts...)
	defer cancel()

	j.listenForNetworkEvents(taskCtx)

	err := chromedp.Run(taskCtx,
		network.Enable(),
		chromedp.Navigate(j.config.URL.String()),
		chromedp.Sleep(j.config.JS.Wait), // Wait for JS to execute, configurable
	)

	if err != nil {
		return nil, err
	}

	j.log.Info().Int("count", len(j.urls)).Msg("JavaScript-based crawl finished")

	var urls []*url.URL
	j.urlsLock.Lock()
	defer j.urlsLock.Unlock()

	for uStr := range j.urls {
		u, err := url.Parse(uStr)
		if err != nil {
			j.log.Warn().Err(err).Msgf("could not parse discovered URL: %s", uStr)
			continue
		}
		urls = append(urls, u)
	}

	return urls, nil
}

func (j *JS) listenForNetworkEvents(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			u, err := url.Parse(ev.Request.URL)
			if err != nil {
				j.log.Warn().Err(err).Msgf("could not parse URL from network event: %s", ev.Request.URL)
				return
			}

			// Apply exclude rules from config
			if !j.config.IsURLAllowed(u) {
				return
			}

			j.urlsLock.Lock()
			j.urls[u.String()] = struct{}{}
			j.urlsLock.Unlock()
		}
	})
}
