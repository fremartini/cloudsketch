package application_gateway

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/types"
)

type handler struct{}

const (
	TYPE   = types.APPLICATION_GATEWAY
	IMAGE  = images.APPLICATION_GATEWAY
	WIDTH  = 64
	HEIGHT = 64
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

	sourceId := (*resource_map)[source.Id].Node.Id()

	for _, target := range targets {
		// don't draw arrows to subnets
		if target.Type == types.SUBNET {
			continue
		}

		targetId := (*resource_map)[target.Id].Node.Id()

		arrows = append(arrows, node.NewArrow(sourceId, targetId, nil))
	}

	return arrows

}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
