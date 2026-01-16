package dalin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/felixdorn/bare/core/domain/analyzer"
	"github.com/felixdorn/bare/core/domain/crawler"
	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules" // Register linting rules
	"github.com/felixdorn/bare/core/domain/reporter"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/felixdorn/bare/core/handler/cli/cli"
	"github.com/spf13/cobra"
)

// Common errors for link filtering
var (
	ErrExternal = errors.New("external link")
	ErrExcluded = errors.New("excluded by pattern")
)

func runReport(c *cli.CLI, cmd *cobra.Command, args []string) error {
	// Get URL from args or flag
	var siteURL *url.URL
	var err error

	if len(args) > 0 {
		urlStr := args[0]
		if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
			urlStr = "http://" + urlStr
		}
		siteURL, err = url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
	} else if cmd.Flags().Changed("url") {
		urlStr, _ := cmd.Flags().GetString("url")
		if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
			urlStr = "http://" + urlStr
		}
		siteURL, err = url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
	} else {
		return fmt.Errorf("URL is required: provide as argument or use --url flag")
	}

	output, _ := cmd.Flags().GetString("output")
	workers, _ := cmd.Flags().GetInt("workers")
	entrypoints, _ := cmd.Flags().GetStringSlice("entrypoint")
	excludes, _ := cmd.Flags().GetStringSlice("exclude")

	// JS config
	jsEnabled, _ := cmd.Flags().GetBool("js-enabled")
	jsWait, _ := cmd.Flags().GetDuration("js-wait")
	jsMaxTabs, _ := cmd.Flags().GetInt("js-max-tabs")
	jsExecutable, _ := cmd.Flags().GetString("js-executable")
	jsFlags, _ := cmd.Flags().GetStringSlice("js-flag")

	log := c.Log()

	// Build exclude patterns
	excludePatterns := make([]url.Path, len(excludes))
	for i, e := range excludes {
		excludePatterns[i] = url.Path(e)
	}
	excludePaths := url.Paths(excludePatterns)

	// Collected page reports
	var pages []reporter.PageReport
	var pagesMu sync.Mutex

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create the appropriate fetcher based on JS config
	var fetcher crawler.Fetcher
	if jsEnabled {
		wait := int(jsWait.Milliseconds())
		if wait == 0 {
			wait = 2000 // default 2 seconds
		}
		maxTabs := jsMaxTabs
		if maxTabs == 0 {
			maxTabs = 1
		}
		jsFetcher, err := crawler.NewJSFetcher(crawler.JSFetcherOptions{
			Wait:           wait,
			MaxTabs:        maxTabs,
			ExecutablePath: jsExecutable,
			Flags:          jsFlags,
			Logger:         log,
		})
		if err != nil {
			return fmt.Errorf("failed to create JS fetcher: %w", err)
		}
		defer jsFetcher.Close()
		fetcher = jsFetcher
	} else {
		fetcher = crawler.NewHTTPFetcher(nil)
	}

	fmt.Printf("Crawling %s...\n", siteURL.String())

	cr := crawler.New(crawler.Config{
		BaseURL:     siteURL,
		WorkerCount: workers,
		Entrypoints: entrypoints,
		Logger:      log,
		Fetcher:     fetcher,

		OnNewLink: func(page *crawler.Page, link crawler.Link) error {
			// Only extract links from crawlable pages (HTML)
			if !isCrawlable(page.URL) {
				return errors.New("source page is not crawlable")
			}

			// Only follow internal links
			if !link.URL.IsInternal(page.URL) {
				return ErrExternal
			}

			// Check excludes
			if excludePaths.MatchAny(link.URL.Path) {
				return ErrExcluded
			}

			return nil
		},

		OnPage: func(page *crawler.Page) {
			// Only report on HTML pages, not assets
			if !isCrawlable(page.URL) {
				return
			}

			// Analyze the page for metadata and images
			analysis, err := analyzer.Analyze(page.Body, page.URL)
			if err != nil {
				log.Error().Err(err).Str("url", page.URL.String()).Msg("Failed to analyze page")
				return
			}

			// Run linting rules
			lints, err := linter.Check(page.Body, page.URL, analysis)
			if err != nil {
				log.Error().Err(err).Str("url", page.URL.String()).Msg("Failed to lint page")
				lints = nil
			}

			pageReport := reporter.PageReport{
				URL:         page.URL.String(),
				Title:       analysis.Title,
				Description: analysis.Description,
				Canonical:   analysis.Canonical,
				StatusCode:  page.StatusCode,
				Images:      analysis.Images,
				Lints:       lints,
			}

			pagesMu.Lock()
			pages = append(pages, pageReport)
			pagesMu.Unlock()

			log.Info().
				Str("url", page.URL.String()).
				Int("images", len(analysis.Images)).
				Int("lints", len(lints)).
				Msg("Analyzed page")
		},
	})

	if err := cr.Run(ctx); err != nil {
		if ctx.Err() != nil {
			fmt.Println("\nCrawl cancelled.")
			return nil
		}
		return fmt.Errorf("crawl failed: %w", err)
	}

	if len(pages) == 0 {
		fmt.Println("No pages found to report.")
		return nil
	}

	// Generate the report
	fmt.Printf("Generating report for %d pages...\n", len(pages))

	rep, err := reporter.New()
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}

	report := &reporter.Report{
		SiteURL:     siteURL.String(),
		GeneratedAt: time.Now(),
		Pages:       pages,
	}

	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := rep.Generate(f, report); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("Report saved to %s\n", output)
	return nil
}

func NewReportCommand(c *cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report [url]",
		Short: "Generate an SEO report for a website",
		Long: `Crawls a website and generates an HTML report with:
- List of all pages with their title, description, and metadata
- Images found on each page with their alt text
- SEO analysis and recommendations (coming soon)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReport(c, cmd, args)
		},
	}

	cmd.Flags().StringP("output", "o", "report.html", "Output file for the HTML report")
	cmd.Flags().String("url", "", "Base URL of the site to analyze")
	cmd.Flags().IntP("workers", "w", 10, "Number of concurrent workers")
	cmd.Flags().StringSlice("entrypoint", []string{"/"}, "Entrypoint paths to seed the crawl")
	cmd.Flags().StringSliceP("exclude", "E", []string{}, "Exclude URLs matching a glob pattern")

	// JS flags
	cmd.Flags().Bool("js-enabled", false, "Enable JavaScript-based crawling for SPAs")
	cmd.Flags().Duration("js-wait", 0, "Time to wait for JS to execute, e.g. 2s, 500ms")
	cmd.Flags().Int("js-max-tabs", 1, "Maximum parallel Chrome tabs for JS fetching")
	cmd.Flags().String("js-executable", "", "Path to Chrome/Chromium executable")
	cmd.Flags().StringSlice("js-flag", []string{}, "Additional Chrome flags")

	return cmd
}

// isCrawlable checks if a URL should be crawled for more links.
func isCrawlable(u *url.URL) bool {
	ext := filepath.Ext(u.Path)
	return ext == "" || ext == ".html"
}
