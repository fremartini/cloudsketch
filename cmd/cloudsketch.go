package cmd

import (
	"cloudsketch/internal/frontends"
	"cloudsketch/internal/frontends/dot"
	"cloudsketch/internal/frontends/drawio"
	"cloudsketch/internal/frontends/drawio/models"
	"cloudsketch/internal/marshall"
	"cloudsketch/internal/providers"
	"cloudsketch/internal/providers/azure"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/urfave/cli/v3"
)

var (
	frontendmap map[string]frontends.Frontend = map[string]frontends.Frontend{
		"drawio": drawio.New(),
		"dot":    dot.New(),
	}
	providermap map[string]providers.Provider = map[string]providers.Provider{
		"azure": azure.NewProvider(),
	}
)

func newCloudsketch(_ context.Context, command *cli.Command) error {
	args := command.Args().Slice()

	if len(args) == 0 {
		return errors.New("command expects one argument")
	}

	fileOrSubscriptionId := args[0]

	f := command.String("frontend")

	p := command.String("provider")

	log.Printf("target frontend is %s\n", f)
	log.Printf("target provider is %s\n", p)

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

		return drawio.New().WriteDiagram(*resources, outFile)
	}

	// otherwise treat it as a subscription id
	subscriptionId := fileOrSubscriptionId

	provider := providermap[p]

	resources, filename, err := provider.FetchResources(subscriptionId)

	if err != nil {
		return err
	}

	frontend := frontendmap[f]

	filename = fmt.Sprintf("%s.%s", filename, f)

	if err := frontend.WriteDiagram(resources, filename); err != nil {
		return err
	}

	// execution succesful. Print the output file name
	log.Print(filename)

	return nil
}
