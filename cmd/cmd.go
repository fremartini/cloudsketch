package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func Execute() {
	name := "cloudsketch"
	cmd := &cli.Command{
		Name:        name,
		Usage:       "Azure to DrawIO",
		UsageText:   fmt.Sprintf("%s <subscription id>", name),
		Description: "convert a Azure subscription to a DrawIO diagram",
		Commands: []*cli.Command{
			newVersion(),
		},
		Action: newCloudsketch,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
