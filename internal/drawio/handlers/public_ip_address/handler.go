package public_ip_address

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.PUBLIC_IP_ADDRESS
	IMAGE  = images.PUBLIC_IP_ADDRESS
	WIDTH  = 65
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

	sourceId := (*resource_map)[source.Id].Node.Id()

	for _, target := range targets {
		targetId := (*resource_map)[target.Id].Node.Id()

		arrows = append(arrows, node.NewArrow(sourceId, targetId, nil))
	}

	return arrows
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
