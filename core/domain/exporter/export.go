package exporter

import (
	"github.com/felixdorn/rera/core/domain/config"
	"github.com/felixdorn/rera/core/domain/httpclient"
	"strings"
	"sync"
)

type LinkType int

const (
	DownloadType LinkType = iota
	CrawlType
)

type QueueItem struct {
	Type  LinkType
	Value any
}

type Export struct {
	ignored map[string]string
	Conf    *config.Config
	baseURL string
	queue   chan QueueItem
}

func NewExport(conf *config.Config) *Export {
	export := &Export{
		ignored: make(map[string]string),
		Conf:    conf,
		baseURL: strings.TrimRight(conf.URL, "/") + "/",
		queue:   make(chan QueueItem),
	}

	return export
}

func (e *Export) run() {
	var wg sync.WaitGroup

	for _, p := range e.Conf.Pages.Crawl {
		e.queue <- QueueItem{CrawlType, p}
	}

	for item := range e.queue {
		wg.Add(1)
		go func(item QueueItem) {
			if item.Type == CrawlType {
				e.exportPage(item.Value.(string))
			} else {
				e.downloadPage(item.Value.(*httpclient.Page))
			}
			wg.Done()
		}(item)
	}

	wg.Wait()
}

func (e *Export) exportPage(path string) error {
	//
}

func (e *Export) downloadPage(page *httpclient.Page) {
	println("here", page.Code)
}
