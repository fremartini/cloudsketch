package virtual_machine

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.VIRTUAL_MACHINE
	IMAGE  = images.VIRTUAL_MACHINE
	WIDTH  = 69
	HEIGHT = 64
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource) *node.Node {
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

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	sourceNode := (*resource_map)[source.Id].Node
	targetNode := (*resource_map)[target.Id].Node

	// if they are in the same group, don't draw the arrow
	if sourceNode.ContainedIn != nil && targetNode.ContainedIn != nil {
		if sourceNode.ContainedIn == targetNode.ContainedIn {
			return nil
		}
	}

	return node.NewArrow(sourceNode.Id(), targetNode.Id())
}

func (*handler) DrawBox(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
