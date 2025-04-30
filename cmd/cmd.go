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
				Validator: func(s string) error {
					validFrontends := []string{"drawio", "dot"}

					valid := list.Contains(validFrontends, func(validFrontend string) bool {
						return s == validFrontend
					})

					if !valid {
						return fmt.Errorf("%s is not a valid frontend. Valid target are %s", s, strings.Join(validFrontends, ","))
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:  "provider",
				Usage: "resource source",
				Value: "azure",
				Validator: func(s string) error {
					validProviders := []string{"azure"}

					valid := list.Contains(validProviders, func(validProvider string) bool {
						return s == validProvider
					})

					if !valid {
						return fmt.Errorf("%s is not a valid provider. Valid target are %s", s, strings.Join(validProviders, ","))
					}

					return nil
				},
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
