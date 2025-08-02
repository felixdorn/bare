package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/handler/cli/cli"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func runServe(c *cli.CLI, cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" && len(args) > 0 {
		dir = args[0]
	}

	if dir == "" {
		conf, err := config.Get()
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			conf = config.NewDefaultConfig()
		}
		dir = conf.Output
	}

	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("directory '%s' does not exist. Run 'bare export' first", dir)
		}
		return err
	}

	port, _ := cmd.Flags().GetInt("port")
	maxRetries := 5
	var ln net.Listener
	var err error
	var addr string

	for i := 0; i < maxRetries; i++ {
		currentPort := port + i
		addr = fmt.Sprintf("127.0.0.1:%d", currentPort)
		ln, err = net.Listen("tcp", addr)
		if err == nil {
			break // Port is available
		}

		if !strings.Contains(err.Error(), "address already in use") {
			// Some other error occurred that we don't know how to handle
			return fmt.Errorf("failed to listen on port %d: %w", currentPort, err)
		}

		// Port is in use
		if i < maxRetries-1 {
			_, _ = fmt.Fprintf(c.Err(), "port %d is in use, trying next port\n", currentPort)
		}
	}

	if err != nil {
		return fmt.Errorf("could not find an available port after %d retries starting from %d", maxRetries, port)
	}
	defer ln.Close()

	url := fmt.Sprintf("http://%s", addr)

	open, _ := cmd.Flags().GetBool("open")
	if open {
		go func() {
			err := browser.OpenURL(url)
			if err != nil {
				_, _ = fmt.Fprintf(c.Err(), "could not open browser: %v\n", err)
			}
		}()
	}

	_, _ = fmt.Fprintf(c.Out(), "Serving static files from %s on %s\n", dir, url)
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	err = http.Serve(ln, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func NewServeCommand(c *cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve [directory]",
		Short: "Serve the exported site",
		Long: `Starts a local web server to preview the contents of the output directory.
By default, it serves the directory specified in bare.toml or 'dist/'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(c, cmd, args)
		},
	}

	cmd.Flags().StringP("dir", "d", "", "Directory to serve from")
	cmd.Flags().IntP("port", "p", 8080, "Port to use")
	cmd.Flags().BoolP("open", "o", false, "Open in browser")

	return cmd
}
