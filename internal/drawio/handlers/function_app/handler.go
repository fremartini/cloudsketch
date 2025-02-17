package function_app

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/types"
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

func (*handler) MapResource(resource *az.Resource) *node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	return node.NewIcon(IMAGE, resource.Name, &geometry)
}

func (*handler) PostProcessIcon(resource *node.ResourceAndNode, resource_map *map[string]*node.ResourceAndNode) *node.Node {
	return nil
}

func (*handler) DrawDependency(source *az.Resource, targets []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	arrows := []*node.Arrow{}

	sourceNode := (*resource_map)[source.Id].Node

	for _, target := range targets {
		// don't draw arrows to subnets
		if target.Type == types.SUBNET {
			continue
		}

		targetNode := (*resource_map)[target.Id].Node

		// if they are in the same group, don't draw the arrow
		if sourceNode.ContainedIn != nil && targetNode.ContainedIn != nil {
			if sourceNode.GetParentOrThis() == targetNode.GetParentOrThis() {
				continue
			}
		}

		arrows = append(arrows, node.NewArrow(sourceNode.Id(), targetNode.Id(), nil))
	}

	// add a dependency to the associated storage account.
	// this property only contains the name of the storage account
	// so it has to be uniquely identified among all storage accounts
	storageAccountName := source.Properties["storageAccountName"]

	resources := []*node.ResourceAndNode{}
	for _, r := range *resource_map {
		resources = append(resources, r)
	}

	resources = list.Filter(resources, func(ran *node.ResourceAndNode) bool {
		return ran.Resource.Type == types.STORAGE_ACCOUNT && strings.Contains(ran.Resource.Name, storageAccountName)
	})

	arrows = append(arrows, node.NewArrow(sourceNode.Id(), resources[0].Node.Id(), nil))

	// add a dependency to the outbound subnet
	dashed := "dashed=1"

	outboundSubnet := source.Properties["outboundSubnet"]
	outboundSubnetNode := (*resource_map)[outboundSubnet].Node
	arrows = append(arrows, node.NewArrow(sourceNode.Id(), outboundSubnetNode.Id(), &dashed))

	return arrows
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
