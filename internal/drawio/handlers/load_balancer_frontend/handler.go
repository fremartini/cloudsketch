package load_balancer_frontend

import (
	"azsample/internal/az"
	"azsample/internal/drawio/handlers/network_interface"
	"azsample/internal/drawio/handlers/node"
)

type handler struct{}

const (
	TYPE   = az.LOAD_BALANCER_FRONTEND
	IMAGE  = network_interface.IMAGE
	WIDTH  = network_interface.WIDTH
	HEIGHT = network_interface.HEIGHT
)

func New() *handler {
	return &handler{}
}

func (*handler) DrawIcon(resource *az.Resource, _ *map[string]*node.ResourceAndNode) []*node.Node {
	properties := node.Properties{
		X:      0,
		Y:      0,
		Width:  WIDTH,
		Height: HEIGHT,
	}

	n := node.NewIcon(IMAGE, resource.Name, &properties)

	return []*node.Node{n}
}

func (*handler) DrawDependency(source, target *az.Resource, nodes *map[string]*node.Node) *node.Arrow {
	sourceId := (*nodes)[source.Id].Id()
	targetId := (*nodes)[target.Id].Id()

	return node.NewArrow(sourceId, targetId)
}

func (*handler) DrawBox(resources []*az.Resource, resource_map *map[string]*node.ResourceAndNode) []*node.Node {
	return []*node.Node{}
}
