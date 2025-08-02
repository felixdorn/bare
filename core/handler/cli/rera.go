package cli

import (
	"github.com/felixdorn/bare/core/handler/cli/cli"
	"github.com/felixdorn/bare/core/handler/cli/cmd"
)

func New(opts ...cli.Opt) *cli.CLI {
	bare := cli.New(
		opts...,
	)

	bare.Add(
		cmd.NewInitCommand(bare),
		cmd.NewExportCommand(bare),
		cmd.NewServeCommand(bare),
	)

	return bare
}
