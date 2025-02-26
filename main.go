package main

import (
	"cloudsketch/internal/drawio"
	"cloudsketch/internal/marshall"
	"cloudsketch/internal/providers/azure"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

var version string

func main() {
	name := "cloudsketch"

	cmd := &cli.Command{
		Name:        name,
		Usage:       "Azure to DrawIO",
		UsageText:   fmt.Sprintf("%s <subscription id>", name),
		Description: "convert a Azure subscription to a DrawIO diagram",
		Commands: []*cli.Command{
			{
				Name:        "version",
				Aliases:     []string{"v"},
				Description: "show version",
				Action: func(_ context.Context, command *cli.Command) error {
					fmt.Println(version)

					return nil
				},
			},
		},
		Action: func(_ context.Context, command *cli.Command) error {
			args := command.Args().Slice()

			if len(args) == 0 {
				return errors.New("command expects one argument")
			}

			fileOrSubscriptionId := args[0]

			// command can either be a subscription id or a file name
			if strings.HasSuffix(fileOrSubscriptionId, ".json") {
				// if the file ends in .json, assume its a valid json file that contains previously populated Azure resources
				file := fileOrSubscriptionId

				log.Printf("using existing file %s\n", file)

				resources, err := marshall.UnmarshallResources(file)

				if err != nil {
					return err
				}

				outFile := strings.ReplaceAll(file, ".json", ".drawio")

				return drawio.New().WriteDiagram(outFile, resources)
			}

			// otherwise treat it as a subscription id
			subscriptionId := fileOrSubscriptionId

			provider := azure.NewProvider()

			resources, filename, err := provider.FetchResources(subscriptionId)

			if err != nil {
				return err
			}

			filename = fmt.Sprintf("%s.drawio", filename)

			if err := drawio.New().WriteDiagram(filename, resources); err != nil {
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
