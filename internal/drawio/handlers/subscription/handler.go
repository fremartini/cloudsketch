package subscription

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/drawio/handlers/node"
	"cloudsketch/internal/drawio/images"
	"cloudsketch/internal/drawio/types"
)

type handler struct{}

const (
	TYPE   = types.SUBSCRIPTION
	IMAGE  = images.SUBSCRIPTION
	WIDTH  = 68
	HEIGHT = 68
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
	return []*node.Arrow{}
}

func (*handler) GroupResources(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
