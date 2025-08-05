package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/domain/exporter"
	"github.com/felixdorn/bare/core/domain/httpclient"
	"github.com/felixdorn/bare/core/domain/js"
	"github.com/felixdorn/bare/core/domain/rewriter"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/felixdorn/bare/core/handler/cli/cli"
	"github.com/spf13/cobra"
)

func runExport(c *cli.CLI, cmd *cobra.Command, args []string) error {
	conf, err := config.Get()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		conf = config.NewDefaultConfig()
	}

	if len(args) > 0 {
		uStr := args[0]
		if !strings.HasPrefix(uStr, "http://") && !strings.HasPrefix(uStr, "https://") {
			uStr = "http://" + uStr
		}

		u, err := url.Parse(uStr)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		conf.URL = u
	}

	output, _ := cmd.Flags().GetString("output")
	if output != "" {
		conf.Output = output
	}

	excludes, _ := cmd.Flags().GetStringSlice("exclude")
	for _, p := range excludes {
		conf.Pages.Exclude = append(conf.Pages.Exclude, url.Path(p))
	}

	withJS, _ := cmd.Flags().GetBool("with-js")
	if !cmd.Flags().Changed("with-js") {
		withJS = conf.JS.Enabled
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var jsCrawler js.Crawler
	if withJS {
		jsCrawler = js.New(conf, c.Log())
	} else {
		jsCrawler = js.NewNoop()
	}

	jsURLs, err := jsCrawler.Run(ctx)
	if err != nil {
		return fmt.Errorf("error during JavaScript crawl: %w", err)
	}

	if len(jsURLs) > 0 {
		log := c.Log()
		log.Info().Int("count", len(jsURLs)).Msg("seeding crawler with URLs discovered via JavaScript")

		// Add discovered URLs to the entrypoints list, avoiding duplicates
		existingEntrypointUrls := make(map[string]struct{})
		for _, u := range conf.Pages.Entrypoints {
			existingEntrypointUrls[string(u)] = struct{}{}
		}

		for _, u := range jsURLs {
			// we only care about the path and query
			p := url.Path(u.RequestURI())
			if _, exists := existingEntrypointUrls[string(p)]; !exists {
				conf.Pages.Entrypoints = append(conf.Pages.Entrypoints, p)
				existingEntrypointUrls[string(p)] = struct{}{}
			}
		}
	}

	client := httpclient.New(nil)
	export := exporter.NewExport(conf, c.Log(), client)
	if err := export.Run(ctx); err != nil {
		return err
	}

	fmt.Println("Rewriting URLs...")
	rw := rewriter.New(conf.Output, conf.URL)
	if err := rw.Run(); err != nil {
		return fmt.Errorf("error rewriting URLs: %w", err)
	}

	return nil
}

func NewExportCommand(c *cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [url]",
		Short: "Export the website",
		Long: `Exports a website by crawling it and saving the pages and assets.
The starting URL can be provided as an argument. If not provided, the URL from bare.toml will be used.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(c, cmd, args)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output directory for the exported site")
	cmd.Flags().Bool("with-js", false, "Enable JavaScript-based crawling to discover more assets")
	cmd.Flags().StringSliceP("exclude", "E", []string{}, "Exclude URLs matching a glob pattern (can be used multiple times)")

	return cmd
}
