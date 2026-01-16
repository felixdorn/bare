package bare

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

	// Optional URL flag override
	if cmd.Flags().Changed("url") {
		uStr, _ := cmd.Flags().GetString("url")
		if !strings.HasPrefix(uStr, "http://") && !strings.HasPrefix(uStr, "https://") {
			uStr = "http://" + uStr
		}
		u, err := url.Parse(uStr)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		conf.URL = u
	}

	// Positional URL argument (takes precedence over flag)
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

	// Output directory
	if cmd.Flags().Changed("output") {
		output, _ := cmd.Flags().GetString("output")
		conf.Output = output
	}

	// Entrypoints
	if cmd.Flags().Changed("entrypoint") {
		eps, _ := cmd.Flags().GetStringSlice("entrypoint")
		conf.Pages.Entrypoints = nil
		for _, p := range eps {
			conf.Pages.Entrypoints = append(conf.Pages.Entrypoints, url.Path(p))
		}
	}

	// Extract-only
	if cmd.Flags().Changed("extract-only") {
		eos, _ := cmd.Flags().GetStringSlice("extract-only")
		conf.Pages.ExtractOnly = nil
		for _, p := range eos {
			conf.Pages.ExtractOnly = append(conf.Pages.ExtractOnly, url.Path(p))
		}
	}

	// Excludes
	if cmd.Flags().Changed("exclude") {
		ex, _ := cmd.Flags().GetStringSlice("exclude")
		conf.Pages.Exclude = nil
		for _, p := range ex {
			conf.Pages.Exclude = append(conf.Pages.Exclude, url.Path(p))
		}
	}

	// Workers
	if cmd.Flags().Changed("workers") {
		workers, _ := cmd.Flags().GetInt("workers")
		conf.WorkersCount = workers
	}

	// JS config
	if cmd.Flags().Changed("js-enabled") {
		enabled, _ := cmd.Flags().GetBool("js-enabled")
		conf.JS.Enabled = enabled
	} else if cmd.Flags().Changed("with-js") { // backward compatibility
		enabled, _ := cmd.Flags().GetBool("with-js")
		conf.JS.Enabled = enabled
	}

	if cmd.Flags().Changed("js-wait") {
		wait, _ := cmd.Flags().GetDuration("js-wait")
		conf.JS.Wait = wait
	}

	if cmd.Flags().Changed("js-executable") {
		exe, _ := cmd.Flags().GetString("js-executable")
		conf.JS.ExecutablePath = exe
	}

	if cmd.Flags().Changed("js-flag") {
		flags, _ := cmd.Flags().GetStringSlice("js-flag")
		conf.JS.Flags = flags
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var jsCrawler js.Crawler
	if conf.JS.Enabled {
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

	export := exporter.NewExport(conf, c.Log(), nil)
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
	cmd.Flags().String("url", "", "Base URL of the site to export (overrides bare.toml)")
	cmd.Flags().IntP("workers", "w", 0, "Number of concurrent workers")
	cmd.Flags().StringSlice("entrypoint", []string{}, "Entrypoint paths to seed the crawl (can be used multiple times)")
	cmd.Flags().StringSliceP("exclude", "E", []string{}, "Exclude URLs matching a glob pattern (can be used multiple times)")
	cmd.Flags().StringSliceP("extract-only", "x", []string{}, "Only extract links from these paths without saving content (can be used multiple times)")
	cmd.Flags().Bool("js-enabled", false, "Enable or disable JavaScript-based crawling")
	cmd.Flags().Bool("with-js", false, "Enable JavaScript-based crawling to discover more assets")
	_ = cmd.Flags().MarkDeprecated("with-js", "use --js-enabled instead")
	cmd.Flags().Duration("js-wait", 0, "Time to wait for JS to execute, e.g. 2s, 500ms")
	cmd.Flags().String("js-executable", "", "Path to Chrome/Chromium executable")
	cmd.Flags().StringSlice("js-flag", []string{}, "Additional Chrome flags (can be used multiple times)")

	return cmd
}
