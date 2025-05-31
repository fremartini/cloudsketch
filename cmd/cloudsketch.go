package cmd

import (
	"cloudsketch/internal/config"
	"cloudsketch/internal/datastructures/build_graph"
	"cloudsketch/internal/frontends"
	"cloudsketch/internal/frontends/dot"
	"cloudsketch/internal/frontends/drawio"
	frontendModels "cloudsketch/internal/frontends/models"
	"cloudsketch/internal/list"
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

	frontend, ok := frontendmap[frontendString]

	if !ok {
		return fmt.Errorf("unknown frontend %s", frontendString)
	}

	providerString := command.String("provider")

	provider, ok := providermap[providerString]

	if !ok {
		return fmt.Errorf("unknown frontend %s", frontendString)
	}

	log.Printf("target frontend is %s\n", frontendString)
	log.Printf("target provider is %s\n", providerString)

	var resources []*providers.Resource
	var filename string

	// command can either be a subscription id or a file name
	if strings.HasSuffix(fileOrSubscriptionId, ".json") {
		// if the file ends in .json, assume its a valid json file that contains previously populated Azure resources
		existingResources, existingFilename, err := useExistingFile(fileOrSubscriptionId, frontendString)

		if err != nil {
			return err
		}

		resources = existingResources
		filename = existingFilename
	} else {
		// otherwise treat it as a subscription id
		existingResources, existingFilename, err := createNewFile(fileOrSubscriptionId, frontendString, provider)

		if err != nil {
			return err
		}

		resources = existingResources
		filename = existingFilename
	}

	frontendResources, err := mapToDomainModels(resources)

	if err != nil {
		return err
	}

	frontendResources = removeBlacklistedResources(frontendResources)

	if err := frontend.WriteDiagram(frontendResources, filename); err != nil {
		return err
	}

	// execution succesful. Print the output file name
	log.Print(filename)

	return nil
}

func removeBlacklistedResources(frontendResources []*frontendModels.Resource) []*frontendModels.Resource {
	config, ok := config.Read()

	if !ok {
		// no config exists. Return all resources
		return frontendResources
	}

	// config is present. Remove all blacklisted resources
	toReturn := list.Filter(frontendResources, func(r *frontendModels.Resource) bool {
		return !list.Contains(config.Blacklist, func(entry string) bool { return entry == r.Type })
	})

	toReturn = list.Map(toReturn, func(r *frontendModels.Resource) *frontendModels.Resource {
		r.DependsOn = list.Filter(r.DependsOn, func(r *frontendModels.Resource) bool {
			return !list.Contains(config.Blacklist, func(entry string) bool { return entry == r.Type })
		})

		return r
	})

	return toReturn
}

func useExistingFile(file, frontendString string) ([]*providers.Resource, string, error) {
	log.Printf("using existing file %s\n", file)

	resources, err := marshall.UnmarshallResources[[]*providers.Resource](file)

	if err != nil {
		return nil, "", err
	}

	outFile := strings.ReplaceAll(file, ".json", fmt.Sprintf(".%s", frontendString))

	return *resources, outFile, nil
}

func createNewFile(subscriptionId, frontendString string, provider providers.Provider) ([]*providers.Resource, string, error) {
	resources, filename, err := provider.FetchResources(subscriptionId)

	if err != nil {
		return nil, "", err
	}

	// cache resources for next run
	err = marshall.MarshallResources(fmt.Sprintf("%s.json", filename), resources)

	filename = fmt.Sprintf("%s.%s", filename, frontendString)

	return resources, filename, err
}

func mapToDomainModels(resources []*providers.Resource) ([]*frontendModels.Resource, error) {
	resource_map := &map[string]*frontendModels.Resource{}

	tasks := list.Map(resources, func(r *providers.Resource) *build_graph.Task {
		return build_graph.NewTask(r.Id, r.DependsOn, []string{}, []string{}, func() { addDependenciesFromIds(r, resource_map) })
	})

	bg, err := build_graph.NewGraph(tasks)

	if err != nil {
		return nil, fmt.Errorf("error during construction of dependency graph: %+v", err)
	}

	for _, task := range tasks {
		bg.Resolve(task)
	}

	domainResources := []*frontendModels.Resource{}
	for _, v := range *resource_map {
		domainResources = append(domainResources, v)
	}

	return domainResources, nil
}

func addDependenciesFromIds(resource *providers.Resource, resource_map *map[string]*frontendModels.Resource) {
	if (*resource_map)[resource.Id] != nil {
		// resource already registered
		return
	}

	dependencies := list.Map(resource.DependsOn, func(d string) *frontendModels.Resource {
		return (*resource_map)[d]
	})

	(*resource_map)[resource.Id] = &frontendModels.Resource{
		Id:         resource.Id,
		Type:       resource.Type,
		Name:       resource.Name,
		DependsOn:  dependencies,
		Properties: resource.Properties,
	}
}
