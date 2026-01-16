package cli

import (
	"github.com/felixdorn/bare/core/handler/cli/cli"
	"github.com/felixdorn/bare/core/handler/cli/dalin"
)

func NewDalin(opts ...cli.Opt) *cli.CLI {
	// Prepend WithName so it can be overridden by user opts
	allOpts := append([]cli.Opt{cli.WithName("dalin")}, opts...)
	app := cli.New(
		allOpts...,
	)

	app.Add(
		dalin.NewReportCommand(app),
	)

	return app
}
