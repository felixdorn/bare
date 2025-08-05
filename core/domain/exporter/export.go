package exporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/domain/httpclient"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/rs/zerolog"
)

// workerResult holds the outcome of a worker's task.
type workerResult struct {
	pageURL   *url.URL
	foundURLs []*url.URL
	err       error
}

// Export manages the crawling and exporting process.
type Export struct {
	Conf    *config.Config
	baseURL *url.URL
	log     zerolog.Logger
	client  *httpclient.Client
}

// NewExport creates a new Export instance.
func NewExport(conf *config.Config, log zerolog.Logger, client *httpclient.Client) *Export {
	return &Export{
		Conf:    conf,
		baseURL: conf.URL,
		log:     log,
		client:  client,
	}
}

// Run executes the export process using a centralized controller model.
func (e *Export) Run(ctx context.Context) error {
	numWorkers := e.Conf.WorkersCount
	e.log.Debug().Msg("Starting export with centralized controller")

	tasksChan := make(chan *url.URL, numWorkers)
	resultsChan := make(chan workerResult, numWorkers)
	var workersWg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go e.worker(ctx, i+1, &workersWg, tasksChan, resultsChan)
	}
	e.log.Debug().Int("workers", numWorkers).Msg("Started workers")

	// Controller state
	queue := make([]*url.URL, 0)
	visited := make(map[string]bool)
	activeWorkers := 0

	fmt.Println(e.Conf.Pages.Entrypoints)
	//
	// Seed the queue with initial entrypoints
	for _, p := range e.Conf.Pages.Entrypoints {
		u, err := url.Parse(string(p))
		if err != nil {
			e.log.Error().Err(err).Str("path", string(p)).Msg("Invalid entrypoint path in config")
			continue
		}
		resolvedURL := e.baseURL.ResolveReference(u)

		if !e.Conf.IsURLAllowed(resolvedURL) {
			e.log.Debug().Str("url", resolvedURL.String()).Msg("Skipping disallowed initial URL")
			continue
		}

		if !visited[resolvedURL.String()] {
			visited[resolvedURL.String()] = true
			queue = append(queue, resolvedURL)
		}
	}
	e.log.Debug().Msg("Initial queue populated")

	// Controller loop
controllerLoop:
	for len(queue) > 0 || activeWorkers > 0 {
		var task *url.URL
		var sendChan chan<- *url.URL

		// Only prepare to send a task if the queue is not empty.
		// This prevents a busy-wait when the queue is empty but workers are still active.
		if len(queue) > 0 {
			task = queue[0]
			sendChan = tasksChan
		}

		select {
		case <-ctx.Done():
			e.log.Info().Msg("Context cancelled, shutting down...")
			// Stop processing, the cleanup logic below will handle the rest.
			break controllerLoop

		case sendChan <- task:
			e.log.Debug().Str("url", task.String()).Msg("Sent task")
			queue = queue[1:]
			activeWorkers++

		case result := <-resultsChan:
			e.log.Debug().Str("url", result.pageURL.String()).Msg("Received result")
			activeWorkers--
			if result.err != nil {
				// Don't log context cancellation errors as failures.
				if ctx.Err() == nil {
					e.log.Error().Err(result.err).Str("url", result.pageURL.String()).Msg("Failed to process URL")
				}
				continue
			}

			// Add new, unvisited, and crawlable URLs to the queue.
			if isCrawlable(result.pageURL) {
				for _, link := range result.foundURLs {
					if !visited[link.String()] {
						// Mark as visited immediately to prevent duplicates in the queue.
						visited[link.String()] = true
						if !e.Conf.IsURLAllowed(link) {
							e.log.Debug().Str("url", link.String()).Msg("Skipping disallowed URL")
							continue
						}
						queue = append(queue, link)
						e.log.Debug().Str("url", link.String()).Msg("Queued new link")
					}
				}
			}
		}
	}

	// Graceful shutdown: wait for active workers to finish and close channels.
	e.log.Debug().Msg("Work finished or cancelled. Starting cleanup.")

	// Drain any remaining results from workers that were in-flight when cancellation happened.
	for activeWorkers > 0 {
		<-resultsChan
		activeWorkers--
		e.log.Debug().Int("remaining_workers", activeWorkers).Msg("Draining worker result.")
	}

	e.log.Debug().Msg("Closing tasks channel.")
	close(tasksChan)

	e.log.Debug().Msg("Waiting for all workers to terminate.")
	workersWg.Wait()

	if ctx.Err() != nil {
		fmt.Println("\nExport cancelled.")
		return nil
	}

	fmt.Println("Export finished.")
	return nil
}

// isCrawlable checks if a URL should be crawled for more links.
func isCrawlable(u *url.URL) bool {
	ext := filepath.Ext(u.Path)
	return ext == "" || ext == ".html"
}

// worker fetches a URL, saves it, and sends the results back to the controller.
func (e *Export) worker(ctx context.Context, id int, wg *sync.WaitGroup, tasks <-chan *url.URL, results chan<- workerResult) {
	defer wg.Done()
	log := e.log.With().Int("worker_id", id).Logger()
	log.Debug().Msg("Worker started")

	for pageURL := range tasks {

		log.Debug().Str("url", pageURL.String()).Msg("Received task")
		page, err := e.client.GetPage(ctx, pageURL)
		if err != nil {
			results <- workerResult{pageURL: pageURL, err: fmt.Errorf("failed to get page: %w", err)}
			continue
		}
		log.Debug().Str("url", pageURL.String()).Msg("Successfully fetched page")

		// If the page is NOT marked as extract-only, save its content.
		if !e.Conf.IsExtractOnly(pageURL) {
			path := pageURL.ToPath(e.Conf.Output)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				results <- workerResult{pageURL: pageURL, err: fmt.Errorf("failed to create directory for %s: %w", path, err)}
				continue
			}

			file, err := os.Create(path)
			if err != nil {
				results <- workerResult{pageURL: pageURL, err: fmt.Errorf("failed to create file %s: %w", path, err)}
				continue
			}

			_, err = io.Copy(file, page.Body)
			_ = file.Close() // Close file regardless of copy error.
			if err != nil {
				results <- workerResult{pageURL: pageURL, err: fmt.Errorf("failed to write file %s: %w", path, err)}
				continue
			}
			log.Info().Str("url", pageURL.String()).Str("path", path).Msg("Exported page")
		} else {
			log.Info().Str("url", pageURL.String()).Msg("Extracting links only, skipping save")
		}

		// Always send the found links back to the controller.
		results <- workerResult{pageURL: pageURL, foundURLs: page.Links}
	}

	log.Debug().Msg("Worker shutting down")
}
