package cli

import (
	"fmt"
	"github.com/felixdorn/rera/core/static"
	"github.com/rs/zerolog"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Stdin provides a minimal interface for  reading stdin.
type Stdin interface {
	io.Reader
	Fd() uintptr
}

// CLI is the command line interface for Name.
type CLI struct {
	out io.Writer
	err io.Writer
	in  Stdin

	log zerolog.Logger

	commands []*cobra.Command
}

// Out returns the current output writer.
func (c *CLI) Out() io.Writer {
	return c.out
}

// In returns the current input reader.
func (c *CLI) In() Stdin {
	return c.in
}

// Err returns the current error writer.
func (c *CLI) Err() io.Writer {
	return c.err
}

// Run the name CLI with the given command line arguments.
// and returns the exit code for the command.
func (c *CLI) Run(args []string) int {
	cli := &cobra.Command{
		Use:           "bin",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: fmt.Sprintf("%s, build %s", static.Version, static.Commit),
	}

	cli.SetHelpCommand(&cobra.Command{
		Use:    "__help",
		Hidden: true,
	})

	cli.AddCommand(c.commands...)

	cli.SetArgs(args)
	cli.SetOut(c.Out())
	cli.SetIn(c.In())
	cli.SetErr(c.Err())

	command := "root"
	if len(args) > 0 {
		command = args[0]
	}

	if err := cli.Execute(); err == nil {
		c.log.
			Info().
			Str("command", command).
			Msg("command ran successfully")

		return 0
	} else if err == nil {
		return 0
	} else if statusErr, ok := err.(*StatusError); ok { //nolint:errorlint // A status error would only exist at the top level of the error chain.
		_, _ = fmt.Fprintf(c.Err(), "Error: %s\n", statusErr.Status)

		c.log.
			Error().
			Str("error", statusErr.Status).
			Str("command", command).
			Int("code", statusErr.StatusCode).
			Msg("command failed")

		return statusErr.StatusCode
	} else {
		_, _ = fmt.Fprintf(c.Err(), "Error: %s\n", err)

		c.log.Error().
			Str("commmand", command).
			Err(err).
			Int("code", 1).
			Msg("command failed")

		return 1
	}
}

// Add adds the given commands to the CLI.
func (c *CLI) Add(commands ...*cobra.Command) *CLI {
	c.commands = append(c.commands, commands...)

	return c
}

// New returns a new CLI with the given standard IO.
func New(opts ...Opt) *CLI {
	cli := &CLI{
		out: os.Stdout,
		err: os.Stderr,
		in:  os.Stdin,
	}

	for _, opt := range opts {
		opt(cli)
	}

	return cli
}

// StatusError is an error type that contains an exit code.
// It is used to exit with a custom exit code.
type StatusError struct {
	Status     string
	StatusCode int
}

// Error implements the error interface.
func (s *StatusError) Error() string {
	return s.Status
}

type Opt func(*CLI)

func WithStdout(out io.Writer) Opt {
	return func(c *CLI) {
		c.out = out
	}
}

func WithStdin(in Stdin) Opt {
	return func(c *CLI) {
		c.in = in
	}
}

func WithStderr(err io.Writer) Opt {
	return func(c *CLI) {
		c.err = err
	}
}

func When(condition bool, opts ...Opt) Opt {
	return func(c *CLI) {
		if condition {
			for _, opt := range opts {
				opt(c)
			}
		}
	}
}
