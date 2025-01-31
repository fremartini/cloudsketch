package data_factory_integration_runtime

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/node"
	"azsample/internal/drawio/handlers/virtual_machine"
	"azsample/internal/drawio/images"
)

type handler struct{}

const (
	TYPE   = az.DATA_FACTORY_INTEGRATION_RUNTIME
	IMAGE  = images.DATA_FACTORY_INTEGRATION_RUNTIME
	WIDTH  = virtual_machine.WIDTH
	HEIGHT = virtual_machine.HEIGHT
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	geometry := node.Geometry{
		X:      0,
		Y:      0,
		Width:  WIDTH / 2,
		Height: HEIGHT / 2,
	}

	n := node.NewIcon(IMAGE, resource.Name, &geometry)

	return []*node.Node{n}
}

func (*handler) DrawDependency(source, target *az.Resource, resource_map *map[string]*node.ResourceAndNode) *node.Arrow {
	sourceId := (*resource_map)[source.Id].Node.Id()
	targetId := (*resource_map)[target.Id].Node.Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(_ *az.Resource, resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
