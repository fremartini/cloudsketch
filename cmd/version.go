package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var version string

func newVersion() *cli.Command {
	return &cli.Command{
		Name:        "version",
		Aliases:     []string{"v"},
		Description: "Show version",
		Action: func(_ context.Context, command *cli.Command) error {
			fmt.Println(version)

			return nil
		},
	}
}
