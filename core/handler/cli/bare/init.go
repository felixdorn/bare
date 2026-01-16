package bare

import (
	"errors"
	"fmt"
	"github.com/felixdorn/bare/core/domain/config"
	"github.com/felixdorn/bare/core/handler/cli/cli"
	"os"

	"github.com/spf13/cobra"
)

func runInit(cli *cli.CLI) error {
	if _, err := os.Stat("bare.toml"); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return fmt.Errorf("bare.toml already exists")
		}

		return fmt.Errorf("bare.toml already exists: %w", err)
	}

	defaultConfigBytes, err := config.NewDefaultConfig().Export()
	if err != nil {
		return err
	}

	err = os.WriteFile("bare.toml", defaultConfigBytes, 0644)
	if err != nil {
		return fmt.Errorf("could not write bare.toml: %w", err)
	}

	return nil
}

func NewInitCommand(cli *cli.CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the default bare.toml config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cli)
		},
	}

	return cmd
}
