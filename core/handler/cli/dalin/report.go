package dalin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/felixdorn/bare/core/domain/analyzer"
	"github.com/felixdorn/bare/core/domain/crawler"
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

	fmt.Printf("Crawling %s...\n", siteURL.String())

	cr := crawler.New(crawler.Config{
		BaseURL:     siteURL,
		WorkerCount: workers,
		Entrypoints: entrypoints,
		Logger:      log,

		OnNewLink: func(page *crawler.Page, link crawler.Link) error {
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
			// Analyze the page for images
			analysis, err := analyzer.Analyze(page.Body, page.URL)
			if err != nil {
				log.Error().Err(err).Str("url", page.URL.String()).Msg("Failed to analyze page")
				return
			}

			pageReport := reporter.PageReport{
				URL:         page.URL.String(),
				Title:       page.Title,
				Description: page.Description,
				Canonical:   page.Canonical,
				StatusCode:  page.StatusCode,
				Images:      analysis.Images,
			}

			pagesMu.Lock()
			pages = append(pages, pageReport)
			pagesMu.Unlock()

			log.Info().
				Str("url", page.URL.String()).
				Int("images", len(analysis.Images)).
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

	return cmd
}
