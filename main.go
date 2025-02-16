package main

import (
	"cloudsketch/internal/drawio"
	"cloudsketch/internal/providers/azure"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	name := "cloudsketch"

	cmd := &cli.Command{
		Name:        name,
		Usage:       "Azure to DrawIO",
		UsageText:   fmt.Sprintf("%s <subscription id>", name),
		Description: "convert a Azure subscription to a DrawIO diagram",
		Action: func(_ context.Context, c *cli.Command) error {
			args := c.Args().Slice()

			if len(args) == 0 {
				return errors.New("missing Azure subscription id")
			}

			subscriptionId := args[0]

			provider := azure.NewProvider()

			resources, filename, err := provider.FetchResources(subscriptionId)

			if err != nil {
				return err
			}

			filename = fmt.Sprintf("%s.drawio", filename)

			err = drawio.New().WriteDiagram(filename, resources)

			if err != nil {
				return err
			}

			// execution succesful. Print the output file name
			log.Print(filename)

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
