package cmd

import (
	"cloudsketch/internal/list"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func Execute() {
	name := "cloudsketch"
	cmd := &cli.Command{
		Name:        name,
		Usage:       "Azure to DrawIO",
		UsageText:   fmt.Sprintf("%s <subscription id>", name),
		Description: "convert a Azure subscription to a DrawIO diagram",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "frontend",
				Usage: "visualization target",
				Value: "drawio",
				Validator: func(frontend string) error {
					return isValidInput([]string{"drawio", "dot"}, frontend)
				},
			},
			&cli.StringFlag{
				Name:  "provider",
				Usage: "resource source",
				Value: "azure",
				Validator: func(provider string) error {
					return isValidInput([]string{"azure"}, provider)
				},
			},
			&cli.BoolFlag{
				Name:  "forceRefresh",
				Usage: "force fetch resources from provider",
				Value: false,
			},
		},
		Commands: []*cli.Command{
			newVersion(),
		},
		Action: newCloudsketch,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func isValidInput(validInputs []string, input string) error {
	valid := list.Contains(validInputs, func(validProvider string) bool {
		return input == validProvider
	})

	if !valid {
		return fmt.Errorf("%s is not a valid value. Valid target are %s", input, strings.Join(validInputs, ","))
	}

	return nil
}
