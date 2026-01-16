package main

import (
	"os"
	"time"

	"github.com/felixdorn/bare/core/handler/cli"
	cliapp "github.com/felixdorn/bare/core/handler/cli/cli"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Logger()

	c := cli.New(cliapp.WithLogger(log))
	os.Exit(c.Run(os.Args[1:]))
}
