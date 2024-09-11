package cli

import (
	"github.com/felixdorn/rera/core/handler/cli/cli"
	"github.com/felixdorn/rera/core/handler/cli/cmd"
)

func New(opts ...cli.Opt) *cli.CLI {
	rera := cli.New(
		opts...,
	)

	rera.Add(
		cmd.NewInitCommand(rera),
	)

	return rera
}
