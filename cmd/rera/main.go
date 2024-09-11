package main

import (
	"github.com/felixdorn/rera/core/handler/cli"
	"os"
)

func main() {
	os.Exit(cli.New().Run(os.Args[1:]))
}
