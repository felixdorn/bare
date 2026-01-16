package cli

import (
	"github.com/felixdorn/bare/core/handler/cli/bare"
	"github.com/felixdorn/bare/core/handler/cli/cli"
)

func New(opts ...cli.Opt) *cli.CLI {
	app := cli.New(
		opts...,
	)

	app.Add(
		bare.NewInitCommand(app),
		bare.NewExportCommand(app),
		bare.NewServeCommand(app),
	)

	return app
}
