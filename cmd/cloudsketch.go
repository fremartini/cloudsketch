package cmd

import (
	"cloudsketch/internal/frontends"
	"cloudsketch/internal/frontends/dot"
	"cloudsketch/internal/frontends/drawio"
	"cloudsketch/internal/frontends/models"
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

	frontendString := command.String("frontend")

	frontend := frontendmap[frontendString]

	providerString := command.String("provider")

	log.Printf("target frontend is %s\n", frontendString)
	log.Printf("target provider is %s\n", providerString)

	// command can either be a subscription id or a file name
	if strings.HasSuffix(fileOrSubscriptionId, ".json") {
		// if the file ends in .json, assume its a valid json file that contains previously populated Azure resources
		return useExistingFile(fileOrSubscriptionId, frontendString, frontend)
	}

	// otherwise treat it as a subscription id
	return createNewFile(fileOrSubscriptionId, providerString, frontendString, frontend)
}

func useExistingFile(file, frontendString string, frontend frontends.Frontend) error {
	log.Printf("using existing file %s\n", file)

	resources, err := marshall.UnmarshallResources[[]*models.Resource](file)

	if err != nil {
		return err
	}

	outFile := strings.ReplaceAll(file, ".json", fmt.Sprintf(".%s", frontendString))

	if err := frontend.WriteDiagram(*resources, outFile); err != nil {
		return err
	}

	// execution succesful. Print the output file name
	log.Print(outFile)

	return nil
}

func createNewFile(subscriptionId, providerString, frontendString string, frontend frontends.Frontend) error {
	provider := providermap[providerString]

	resources, filename, err := provider.FetchResources(subscriptionId)

	if err != nil {
		return err
	}

	filename = fmt.Sprintf("%s.%s", filename, frontendString)

	if err := frontend.WriteDiagram(resources, filename); err != nil {
		return err
	}

	// execution succesful. Print the output file name
	log.Print(filename)

	return nil
}
