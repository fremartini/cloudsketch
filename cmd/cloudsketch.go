package cmd

import (
	"cloudsketch/internal/frontends/drawio"
	"cloudsketch/internal/frontends/drawio/models"
	"cloudsketch/internal/marshall"
	"cloudsketch/internal/providers/azure"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/urfave/cli/v3"
)

func newCloudsketch(_ context.Context, command *cli.Command) error {
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

		resources, err := marshall.UnmarshallResources[[]*models.Resource](file)

		if err != nil {
			return err
		}

		outFile := strings.ReplaceAll(file, ".json", ".drawio")

		return drawio.New(*resources).WriteDiagram(outFile)
	}

	// otherwise treat it as a subscription id
	subscriptionId := fileOrSubscriptionId

	provider := azure.NewProvider()

	resources, filename, err := provider.FetchResources(subscriptionId)

	if err != nil {
		return err
	}
	/*
		filename = fmt.Sprintf("%s.dot", filename)

		if err := dot.New(resources).WriteDiagram(filename); err != nil {
			return err
		}*/

	filename = fmt.Sprintf("%s.drawio", filename)

	if err := drawio.New(resources).WriteDiagram(filename); err != nil {
		return err
	}

	// execution succesful. Print the output file name
	log.Print(filename)

	return nil
}
