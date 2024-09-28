package cmd

import (
	"github.com/felixdorn/rera/core/domain/config"
	"github.com/felixdorn/rera/core/domain/exporter"
	"github.com/felixdorn/rera/core/domain/httpclient"
	"github.com/felixdorn/rera/core/handler/cli/cli"
	"github.com/spf13/cobra"
)

type Result struct {
	Page  *httpclient.Page
	Error error
}

func runExport(cli *cli.CLI) error {
	conf, err := config.Get()
	if err != nil {
		return err
	}

	exporter.NewExport(conf).Run()

	return nil
}

func NewExportCommand(cli *cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the website",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(cli)
		},
	}

	return cmd
}
