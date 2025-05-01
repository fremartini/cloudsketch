package function_app

import (
	"cloudsketch/internal/frontends/drawio/handlers/node"
	"cloudsketch/internal/frontends/drawio/images"
	"cloudsketch/internal/frontends/models"
	"cloudsketch/internal/frontends/types"
	"cloudsketch/internal/list"
	"strings"
)

type handler struct{}

const (
	TYPE   = types.FUNCTION_APP
	IMAGE  = images.FUNCTION_APP
	WIDTH  = 67
	HEIGHT = 52
)

func New() *handler {
	return &handler{}
}

func (*handler) MapResource(resource *models.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	link := resource.GetLinkOrDefault()

	return node.NewIcon(IMAGE, resource.Name, &geometry, link)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependencies(source *models.Resource, targets []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	arrows := node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{})

	arrows = append(arrows, addDependencyToAssociatedStorageAccount(source, resource_map)...)

	arrows = append(arrows, addDependencyToOutboundSubnet(source, resource_map)...)

	return arrows
}

func addDependencyToAssociatedStorageAccount(source *models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	// this property only contains the name of the storage account
	// so it has to be uniquely identified among all storage accounts
	storageAccountName, ok := source.Properties["storageAccountName"]

	if !ok {
		return []*node.Arrow{}
	}

	resources := []*node.ResourceAndNode{}
	for _, r := range *resource_map {
		resources = append(resources, r)
	}

	resources = list.Filter(resources, func(ran *node.ResourceAndNode) bool {
		return ran.Resource.Type == types.STORAGE_ACCOUNT && strings.Contains(ran.Resource.Name, storageAccountName[0])
	})

	sourceNode := (*resource_map)[source.Id].Node

	return []*node.Arrow{node.NewArrow(sourceNode.Id(), resources[0].Node.Id(), nil)}
}

func addDependencyToOutboundSubnet(source *models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	dashed := "dashed=1"

	outboundSubnet, ok := source.Properties["outboundSubnet"]

	if !ok {
		return []*node.Arrow{}
	}

	outboundSubnetNode := (*resource_map)[outboundSubnet[0]].Node

	sourceNode := (*resource_map)[source.Id].Node

	return []*node.Arrow{node.NewArrow(sourceNode.Id(), outboundSubnetNode.Id(), &dashed)}
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
