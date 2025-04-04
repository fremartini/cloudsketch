package logic_app

import (
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/models"
	"cloudsketch/internal/drawio/types"
)

type handler struct{}

const (
	TYPE   = types.LOGIC_APP
	IMAGE  = images.LOGIC_APP
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
	arrows := node.DrawDependencyArrowsToTarget(source, targets, resource_map, []string{types.SUBNET})

	arrows = append(arrows, addDependencyToOutboundSubnet(source, resource_map)...)

	return arrows
}

func addDependencyToOutboundSubnet(source *models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Arrow {
	dashed := "dashed=1"

	outboundSubnet := source.Properties["outboundSubnet"].(string)
	outboundSubnetNode := (*resource_map)[outboundSubnet].Node

	sourceNode := (*resource_map)[source.Id].Node

	return []*node.Arrow{node.NewArrow(sourceNode.Id(), outboundSubnetNode.Id(), &dashed)}
}

func (*handler) GroupResources(_ *models.Resource, resources []*models.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
